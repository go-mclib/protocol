package javaproto

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/go-mclib/protocol/crypto"
)

// Conn wraps a net.Conn with buffered reads, optional AES/CFB8 encryption and
// packet framing. It implements io.Reader and io.Writer, transparently handling
// encryption/decryption when enabled.
//
// Packet writes and mode switches (compression threshold, encryption) are
// serialized, so WritePacket and WriteWirePacket are safe for concurrent use.
// Reads, and EnableEncryption, must come from a single goroutine.
type Conn struct {
	conn net.Conn
	r    *bufio.Reader

	// decrypt is owned by the reading goroutine; encrypt is guarded by writeMu.
	decrypt   cipher.Stream
	encrypt   cipher.Stream
	encrypted atomic.Bool

	writeMu              sync.Mutex
	compressionThreshold atomic.Int32
}

// NewConn creates a new Conn wrapping the given net.Conn, with compression disabled.
func NewConn(conn net.Conn) *Conn {
	c := &Conn{conn: conn, r: bufio.NewReader(conn)}
	c.compressionThreshold.Store(-1)
	return c
}

// Read implements io.Reader. If encryption is enabled, data is decrypted.
// Bytes are decrypted when consumed rather than when buffered, so enabling
// encryption takes effect exactly at the next unread byte.
func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 && c.decrypt != nil {
		c.decrypt.XORKeyStream(p[:n], p[:n])
	}
	return n, err
}

// Write implements io.Writer. If encryption is enabled, an encrypted copy of p
// is written; p itself is not modified.
func (c *Conn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.encrypt == nil {
		return c.conn.Write(p)
	}
	encrypted := make([]byte, len(p))
	c.encrypt.XORKeyStream(encrypted, p)
	return c.conn.Write(encrypted)
}

// writeFrame writes a frame owned by the caller, encrypting it in place.
// The caller must hold writeMu.
func (c *Conn) writeFrame(frame []byte) error {
	if c.encrypt != nil {
		c.encrypt.XORKeyStream(frame, frame)
	}
	_, err := c.conn.Write(frame)
	return err
}

// WritePacket serializes p and writes it using the current compression threshold.
// Safe for concurrent use from multiple goroutines.
func (c *Conn) WritePacket(p Packet) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// serialize at the transmission boundary: packet writers may validate a
	// session/action context that was canceled while another write held the lock
	wire, err := ToWire(p)
	if err != nil {
		return fmt.Errorf("failed to serialize packet: %w", err)
	}
	if err := c.writeFrame(wire.frame(int(c.compressionThreshold.Load()))); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}
	return nil
}

// WriteWirePacket writes a raw WirePacket using the current compression threshold.
// Safe for concurrent use from multiple goroutines.
func (c *Conn) WriteWirePacket(pkt *WirePacket) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := c.writeFrame(pkt.frame(int(c.compressionThreshold.Load()))); err != nil {
		return fmt.Errorf("failed to write wire packet: %w", err)
	}
	return nil
}

// ReadWirePacket reads a raw WirePacket using the current compression threshold.
func (c *Conn) ReadWirePacket() (*WirePacket, error) {
	return ReadWirePacketFrom(c, int(c.compressionThreshold.Load()))
}

// SetCompressionThreshold sets the compression threshold; -1 disables compression.
// It waits for any in-flight packet write, so no frame mixes two modes.
func (c *Conn) SetCompressionThreshold(threshold int) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.compressionThreshold.Store(int32(threshold))
}

// EnableEncryption enables AES/CFB8 encryption with the 16-byte shared secret
// (see crypto.NewSharedSecret). Call it from the reading goroutine: it applies
// from the next unread byte. It waits for any in-flight packet write, so no
// frame is partially encrypted.
func (c *Conn) EnableEncryption(sharedSecret []byte) error {
	// the secret is both the AES-128 key and the CFB8 IV
	if len(sharedSecret) != 16 {
		return fmt.Errorf("shared secret must be 16 bytes, got %d", len(sharedSecret))
	}
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.encrypt = crypto.NewCFB8(block, sharedSecret, false)
	c.decrypt = crypto.NewCFB8(block, sharedSecret, true)
	c.encrypted.Store(true)
	return nil
}

// Encrypted reports whether encryption is enabled. Safe for concurrent use.
func (c *Conn) Encrypted() bool {
	return c.encrypted.Load()
}

// Close closes the underlying connection.
func (c *Conn) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	if c.conn != nil {
		return c.conn.RemoteAddr()
	}
	return nil
}
