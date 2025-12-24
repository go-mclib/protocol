package java_protocol

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	ns "github.com/go-mclib/protocol/net_structures"
)

type TCPClient struct {
	*BaseTCP
	state                State
	compressionThreshold int
}

func NewTCPClient() *TCPClient {
	return &TCPClient{
		BaseTCP:              NewBaseTCP(nil),
		state:                StateHandshake,
		compressionThreshold: -1,
	}
}

func (c *TCPClient) SetState(state State) {
	c.state = state
}

func (c *TCPClient) GetState() State {
	return c.state
}

func (c *TCPClient) SetCompressionThreshold(threshold int) {
	c.compressionThreshold = threshold
}

func (c *TCPClient) WritePacket(packet *Packet) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	data, err := packet.ToBytes(c.compressionThreshold)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	c.debugf("-> send: state=%v bound=%v id=0x%02X len=%d (pre-encrypt) bytes=%s", packet.State, packet.Bound, int(packet.PacketID), len(data), hexSnippet(data, 256))

	if c.encryption.IsEnabled() {
		enc := c.encryption.Encrypt(data)
		c.debugf("-> send: encrypted len=%d bytes=%s", len(enc), hexSnippet(enc, 256))
		data = enc
	}

	n, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}
	c.debugf("-> send: wrote=%d bytes", n)

	return nil
}

func (c *TCPClient) ReadPacket() (*Packet, error) {
	c.debugf("<- recv: waiting for packet length varint")
	packetLength, err := c.readVarInt()
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %w", err)
	}
	c.debugf("<- recv: length=%d", int(packetLength))

	data := make([]byte, packetLength)
	n, err := io.ReadFull(c.conn, data)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet data: %w", err)
	}
	c.debugf("<- recv: read=%d bytes (encrypted? %v) bytes=%s", n, c.encryption.IsEnabled(), hexSnippet(data, 256))

	if c.encryption.IsEnabled() {
		dec := c.encryption.Decrypt(data)
		c.debugf("<- recv: decrypted len=%d bytes=%s", len(dec), hexSnippet(dec, 256))
		data = dec
	}

	reader := bytes.NewReader(data)

	if c.compressionThreshold >= 0 {
		c.debugf("<- recv: compression enabled (compressionThreshold=%d)", c.compressionThreshold)
		return c.readCompressedPacket(reader)
	}

	c.debugf("<- recv: compression disabled")
	return c.readUncompressedPacket(reader)
}

func (c *TCPClient) readUncompressedPacket(reader *bytes.Reader) (*Packet, error) {
	startLen := reader.Len()
	packetID, err := c.readVarIntFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}

	remainingData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read remaining data: %w", err)
	}
	c.debugf("<- recv: uncompressed id=0x%02X id_bytes=%d data_len=%d total_len=%d", int(packetID), startLen-reader.Len()-len(remainingData), len(remainingData), int(packetID)+len(remainingData))

	// return a generic packet with raw data
	return &Packet{
		State:    c.state,
		Bound:    S2C,
		PacketID: packetID,
		Data:     ns.ByteArray(remainingData),
	}, nil
}

func (c *TCPClient) readCompressedPacket(reader *bytes.Reader) (*Packet, error) {
	before := reader.Len()
	dataLength, err := c.readVarIntFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}

	if dataLength == 0 {
		c.debugf("<- recv: compressed framing with dataLen=0 (actually uncompressed)")
		return c.readUncompressedPacket(reader)
	}

	compressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %w", err)
	}
	c.debugf("<- recv: compressed payload read=%d (declared uncompressed=%d) bytes=%s", len(compressedData), int(dataLength), hexSnippet(compressedData, 256))

	zlibReader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zlibReader.Close()

	uncompressedData, err := io.ReadAll(zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}
	if int(dataLength) != len(uncompressedData) {
		c.debugf("<- recv: WARN uncompressed length mismatch: declared=%d actual=%d (frameVarIntBytes=%d)", int(dataLength), len(uncompressedData), before-reader.Len()-len(compressedData))
	}

	uncompressedReader := bytes.NewReader(uncompressedData)
	packetID, err := c.readVarIntFromReader(uncompressedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}

	remainingData, err := io.ReadAll(uncompressedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read remaining data: %w", err)
	}
	c.debugf("<- recv: compressed id=0x%02X data_len=%d", int(packetID), len(remainingData))

	// return a generic packet with raw data
	return &Packet{
		State:    c.state,
		Bound:    S2C,
		PacketID: packetID,
		Data:     ns.ByteArray(remainingData),
	}, nil
}

func (c *TCPClient) readVarInt() (ns.VarInt, error) {
	var value int32
	var position int
	var currentByte byte

	for {
		buf := make([]byte, 1)
		n, err := io.ReadFull(c.conn, buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				c.debugf("<- recv: EOF while reading VarInt at position=%d bytes_read=%d", position, n)
			}
			return 0, err
		}

		if c.encryption.IsEnabled() {
			buf = c.encryption.Decrypt(buf)
		}
		c.debugf("<- recv: varint byte=0x%02X", buf[0])

		currentByte = buf[0]
		value |= (int32(currentByte) & 0x7F) << position

		if (currentByte & 0x80) == 0 {
			break
		}

		position += 7

		if position >= 32 {
			return 0, fmt.Errorf("VarInt is too big")
		}
	}

	return ns.VarInt(value), nil
}

func (c *TCPClient) readVarIntFromReader(reader *bytes.Reader) (ns.VarInt, error) {
	var value int32
	var position int
	var currentByte byte

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		currentByte = b
		value |= (int32(currentByte) & 0x7F) << position

		if (currentByte & 0x80) == 0 {
			break
		}

		position += 7

		if position >= 32 {
			return 0, fmt.Errorf("VarInt is too big")
		}
	}

	return ns.VarInt(value), nil
}
