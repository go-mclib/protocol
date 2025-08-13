// The `java_protocol` package contains the core structs and functions for working with the Java Edition protocol.
//
// > The Minecraft server accepts connections from TCP clients and communicates with them using packets.
// A packet is a sequence of bytes sent over the TCP connection (note: see `net_structures.ByteArray`).
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
package java_protocol

import (
	"bytes"
	"compress/zlib"
	"fmt"

	ns "github.com/go-mclib/protocol/net_structures"
)

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

// The top-level packet struct.
//
// This is the base struct for all packets.
// It contains the phase, bound, packet ID, and data, as Go structs.
type Packet struct {
	// The state of the packet (handshake, status, login, configuration, play), not sent over network
	State State
	// The direction of the packet, not sent over network
	Bound Bound
	// The ID of the packet, represented as a `VarInt`.
	PacketID ns.VarInt
	// The raw body bytes of the packet (already-encoded fields minus the Packet ID)
	Data ns.ByteArray
}

// NewPacket creates a packet template with no body data.
func NewPacket(state State, bound Bound, packetID ns.VarInt) *Packet {
	return &Packet{
		State:    state,
		Bound:    bound,
		PacketID: packetID,
		Data:     nil,
	}
}

// WithData marshals the provided packet data struct into bytes
// and returns the original packet with the data bytes set.
func (p *Packet) WithData(v any) (*Packet, error) {
	if p == nil {
		return nil, fmt.Errorf("nil packet template")
	}
	dataBytes, err := PacketDataToBytes(v)
	if err != nil {
		return nil, err
	}
	p.Data = dataBytes
	return p, nil
}

// ToBytes marshals the packet into a byte array that can be sent over the network.
//
// If `compressionThreshold` is non-negative, compression is enabled.
// The format of the raw packet varies, depending on compression.
//
// > Once a Set Compression packet (with a non-negative threshold) is sent, zlib compression
// is enabled for all following packets. The format of a packet changes slightly to include
// the size of the uncompressed packet. For serverbound packets, the uncompressed length of
// (Packet ID + Data) must not be greater than 223 or 8388608 bytes. Note that a length equal
// to 223 is permitted, which differs from the compressed length limit. The vanilla client, on
// the other hand, has no limit for the uncompressed length of incoming compressed packets.
//
// > If the size of the buffer containing the packet data and ID (as a VarInt) is smaller than the
// threshold specified in the packet Set Compression. It will be sent as uncompressed.
// This is done by setting the data length as 0. (Comparable to sending a non-compressed format
// with an extra 0 between the length, and packet data).
//
// > If it's larger than or equal to the threshold, then it follows the regular compressed protocol format.
//
// > The vanilla server (but not client) rejects compressed packets smaller than the threshold.
// > Uncompressed packets exceeding the threshold, however, are accepted.
//
// > Compression can be disabled by sending the packet Set Compression with a negative Threshold,
// > or not sending the Set Compression packet at all."
//
// See https://minecraft.wiki/w/Java_Edition_protocol/Packets#Packet_format
func (p *Packet) ToBytes(compressionThreshold int) (ns.ByteArray, error) {
	if compressionThreshold >= 0 {
		return p.toBytesCompressed(compressionThreshold)
	} else {
		return p.toBytesUncompressed()
	}
}

// Structure:
//
//	if (size >= networkCompressionThreshold)
//		packetLength: VarInt(Length of (Data Length) + length of compressed (Packet ID + Data)) +
//		dataLength: VarInt(Length of uncompressed (Packet ID + Data)) +
//		packetID: compressed(VarInt(Packet ID)) +
//		data: compressed(Data)
//	if (size < networkCompressionThreshold)
//		packetLength: VarInt(Length of (Data Length) + length of uncompressed (Packet ID + Data)) +
//		dataLength: VarInt(0) + // compressed data length is 0, which means no compression is used
//		packetID: VarInt(Packet ID) +
//		data: ByteArray(Data)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#With_compression
func (p *Packet) toBytesCompressed(compressionThreshold int) (ns.ByteArray, error) {
	// marshal packet ID and use raw body bytes
	packetIDBytes, err := p.PacketID.ToBytes()
	if err != nil {
		return nil, err
	}
	uncompressedPayload := append(packetIDBytes, p.Data...)
	uncompressedLength := len(uncompressedPayload)

	// threshold check
	if uncompressedLength >= compressionThreshold {
		// compress the payload (marshalled packet ID + marshalled data)
		compressedPayload := compressZlib(uncompressedPayload)

		// build packet
		dataLengthBytes, err := ns.VarInt(uncompressedLength).ToBytes()
		if err != nil {
			return nil, err
		}
		packetContent := append(dataLengthBytes, compressedPayload...)
		packetLengthBytes, err := ns.VarInt(len(packetContent)).ToBytes()
		if err != nil {
			return nil, err
		}

		return append(packetLengthBytes, packetContent...), nil
	} else {
		// uncompressed
		dataLengthBytes, err := ns.VarInt(0).ToBytes() // 0 indicates uncompressed
		if err != nil {
			return nil, err
		}
		packetContent := append(dataLengthBytes, uncompressedPayload...)
		packetLengthBytes, err := ns.VarInt(len(packetContent)).ToBytes()
		if err != nil {
			return nil, err
		}

		return append(packetLengthBytes, packetContent...), nil
	}
}

// Structure:
//
//	packetLength: VarInt(Length of Packet ID + Data) +
//	packetID: VarInt(Packet ID) +
//	data: ByteArray(Data)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Without_compression
func (p *Packet) toBytesUncompressed() (ns.ByteArray, error) {
	// marshal packet ID and use raw body bytes
	packetIDBytes, err := p.PacketID.ToBytes()
	if err != nil {
		return nil, err
	}

	// build packet
	payload := append(packetIDBytes, p.Data...)
	packetLengthBytes, err := ns.VarInt(len(payload)).ToBytes()
	if err != nil {
		return nil, err
	}

	return append(packetLengthBytes, payload...), nil
}

func compressZlib(data []byte) []byte {
	compressedData := bytes.NewBuffer(nil)
	writer := zlib.NewWriter(compressedData)
	writer.Write(data)
	writer.Close()
	return compressedData.Bytes()
}
