// Package javaproto contains the core structs and functions for working with the Java Edition protocol.
//
// > The Minecraft server accepts connections from TCP clients and communicates with them using packets.
// A packet is a sequence of bytes sent over the TCP connection (note: see `wire.ByteArray`).
// The meaning of a packet depends both on its packet ID and the current state of the connection
// (note: each state has its own packet ID counter, so packets in different states can have the same packet ID).
// The initial state of each connection is Handshaking, and state is switched using the packets 'Handshake' and 'Login Success'."
//
// Packet format:
//
// > Packets cannot be larger than (2^21) − 1 or 2 097 151 bytes (the maximum that can be sent in a 3-byte VarInt).
// Moreover, the length field must not be longer than 3 bytes, even if the encoded value is within the limit.
// Unnecessarily long encodings at 3 bytes or below are still allowed.
// For compressed packets, this applies to the Packet Length field, i. e. the compressed length.
//
// See https://minecraft.wiki/w/Java_Edition_protocol/Packets
package javaproto

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// Packet is the interface that all typed packet implementations must satisfy.
// Each packet knows its ID, protocol state, and direction.
type Packet interface {
	// ID returns the packet ID for this packet type.
	ID() ns.VarInt
	// State returns the protocol state this packet belongs to.
	State() State
	// Bound returns the direction of this packet (C2S or S2C).
	Bound() Bound
	// Read deserializes the packet data from the buffer.
	Read(buf *ns.PacketBuffer) error
	// Write serializes the packet data to the buffer.
	Write(buf *ns.PacketBuffer) error
}

// State is the phase that the packet is in (handshake, status, login, configuration, play).
// This is not sent over network (server and client automatically transition phases).
type State uint8

const (
	StateHandshake State = iota
	StateStatus
	StateLogin
	StateConfiguration
	StatePlay
)

// String returns the state's name, e.g. "Play".
func (s State) String() string {
	switch s {
	case StateHandshake:
		return "Handshake"
	case StateStatus:
		return "Status"
	case StateLogin:
		return "Login"
	case StateConfiguration:
		return "Configuration"
	case StatePlay:
		return "Play"
	default:
		return fmt.Sprintf("State(%d)", uint8(s))
	}
}

// Bound is the direction that the packet is going.
//
// Serverbound: Client -> Server (C2S)
//
// Clientbound: Server -> Client (S2C)
type Bound uint8

const (
	// Client -> Server (C2S, serverbound)
	C2S Bound = iota
	// Server -> Client (S2C, clientbound)
	S2C
)

const (
	// maxPacketLength is the largest frame vanilla accepts (a 3-byte VarInt).
	maxPacketLength = 1<<21 - 1
	// maxUncompressedLength mirrors vanilla's CompressionDecoder.MAXIMUM_UNCOMPRESSED_LENGTH.
	maxUncompressedLength = 1 << 23
)

// WirePacket represents the raw packet as it appears on the wire.
// It contains only wire-level data without typed field information.
type WirePacket struct {
	// PacketID is the packet identifier.
	PacketID ns.VarInt
	// Data is the raw payload bytes (without the packet ID).
	Data ns.ByteArray
}

// ReadWirePacketFrom reads a WirePacket from the given reader.
// Handles both compressed and uncompressed packet formats based on compressionThreshold.
// Use compressionThreshold < 0 to disable compression.
func ReadWirePacketFrom(r io.Reader, compressionThreshold int) (*WirePacket, error) {
	packetLength, err := ns.DecodeVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %w", err)
	}
	if packetLength < 0 || packetLength > maxPacketLength {
		return nil, fmt.Errorf("invalid packet length %d", packetLength)
	}

	payload := make([]byte, packetLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("failed to read packet data: %w", err)
	}

	if compressionThreshold >= 0 {
		if payload, err = decompressPayload(payload); err != nil {
			return nil, err
		}
	}
	return parsePayload(payload)
}

// WriteTo writes the WirePacket to the given writer.
// Handles both compressed and uncompressed packet formats based on compressionThreshold.
// Use compressionThreshold < 0 to disable compression.
//
// Compression behavior (per Minecraft protocol):
//   - If size >= threshold: packet is zlib compressed
//   - If size < threshold: packet is sent uncompressed (with Data Length = 0)
//   - The vanilla server rejects compressed packets smaller than the threshold
//
// See https://minecraft.wiki/w/Java_Edition_protocol/Packets#Packet_format
func (w *WirePacket) WriteTo(writer io.Writer, compressionThreshold int) error {
	_, err := writer.Write(w.frame(compressionThreshold))
	return err
}

// ReadInto deserializes the wire packet's raw data into a typed Packet.
// Returns an error if the packet ID doesn't match.
func (w *WirePacket) ReadInto(p Packet) error {
	if w == nil {
		return errors.New("nil wire packet")
	}
	if w.PacketID != p.ID() {
		return fmt.Errorf("packet ID mismatch: expected 0x%02X, got 0x%02X", p.ID(), w.PacketID)
	}
	buf := ns.NewReader(w.Data)
	return p.Read(buf)
}

// ReadPacket deserializes a WirePacket into a typed Packet using generics.
// This provides type-safe packet reading without manual type assertions.
//
// Example:
//
//	wire, _ := client.ReadWirePacket()
//	login, err := ReadPacket[LoginSuccessPacket](wire)
func ReadPacket[T any, PT interface {
	*T
	Packet
}](wire *WirePacket) (PT, error) {
	p := new(T)
	pt := PT(p)
	if err := wire.ReadInto(pt); err != nil {
		return nil, err
	}
	return pt, nil
}

// ToWire converts a typed Packet to a WirePacket by serializing its data.
// The resulting WirePacket can then be written to a connection via WriteTo().
func ToWire(p Packet) (*WirePacket, error) {
	buf := ns.NewWriter()
	if err := p.Write(buf); err != nil {
		return nil, fmt.Errorf("failed to serialize packet data: %w", err)
	}
	return &WirePacket{
		PacketID: p.ID(),
		Data:     buf.Bytes(),
	}, nil
}

// parsePayload splits an uncompressed (Packet ID + Data) payload.
func parsePayload(payload []byte) (*WirePacket, error) {
	r := bytes.NewReader(payload)
	packetID, err := ns.DecodeVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}
	return &WirePacket{PacketID: packetID, Data: payload[len(payload)-r.Len():]}, nil
}

// decompressPayload unwraps a compressed-format frame body (Data Length + payload).
// The declared length bounds decompression, so a zlib bomb cannot inflate past it.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#With_compression
func decompressPayload(frame []byte) ([]byte, error) {
	r := bytes.NewReader(frame)
	dataLength, err := ns.DecodeVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}
	// dataLength == 0 means uncompressed despite compression being enabled
	if dataLength == 0 {
		return frame[len(frame)-r.Len():], nil
	}
	if dataLength < 0 || dataLength > maxUncompressedLength {
		return nil, fmt.Errorf("invalid uncompressed data length %d", dataLength)
	}

	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}
	defer zr.Close()

	payload := make([]byte, dataLength)
	if _, err := io.ReadFull(zr, payload); err != nil {
		return nil, fmt.Errorf("failed to decompress %d bytes: %w", dataLength, err)
	}
	var extra [1]byte
	switch _, err := io.ReadFull(zr, extra[:]); {
	case err == nil:
		return nil, fmt.Errorf("decompressed data exceeds declared length %d", dataLength)
	case !errors.Is(err, io.EOF):
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}
	return payload, nil
}

// frame serializes the packet with its length prefix, compressing per threshold
// (< 0 disables compression).
//
// Without compression:
//
//	VarInt(len(Packet ID + Data)) + VarInt(Packet ID) + Data
//
// With compression, below the threshold:
//
//	VarInt(len(Data Length + Packet ID + Data)) + VarInt(0) + VarInt(Packet ID) + Data
//
// With compression, at or above the threshold:
//
//	VarInt(len(Data Length + compressed)) + VarInt(len(Packet ID + Data)) + zlib(VarInt(Packet ID) + Data)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Packet_format
func (w *WirePacket) frame(compressionThreshold int) []byte {
	size := w.PacketID.Len() + len(w.Data)
	switch {
	case compressionThreshold < 0:
		return w.appendPayload(ns.VarInt(size).Append(make([]byte, 0, 5+size)))
	case size < compressionThreshold:
		out := ns.VarInt(size + 1).Append(make([]byte, 0, 6+size))
		return w.appendPayload(append(out, 0))
	}
	compressed := compressZlib(w.appendPayload(make([]byte, 0, size)))
	dataLength := ns.VarInt(size)
	out := ns.VarInt(dataLength.Len() + len(compressed)).Append(make([]byte, 0, 10+len(compressed)))
	return append(dataLength.Append(out), compressed...)
}

func (w *WirePacket) appendPayload(dst []byte) []byte {
	return append(w.PacketID.Append(dst), w.Data...)
}

func compressZlib(data []byte) []byte {
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	_, _ = writer.Write(data) // writes to a bytes.Buffer cannot fail
	_ = writer.Close()
	return compressed.Bytes()
}
