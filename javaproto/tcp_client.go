package javaproto

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

var errNilConn = errors.New("connection is nil")

// TCPClient is a Minecraft protocol client connection.
type TCPClient struct {
	conn   *Conn
	state  atomic.Uint32
	logger *log.Logger // nil disables debug logging
}

// NewTCPClient creates a new TCP client in the handshake state.
func NewTCPClient() *TCPClient {
	return &TCPClient{}
}

// Connect connects to a Minecraft server.
// The address can be "host", "host:port", or will use SRV records if no port specified.
// ctx bounds the SRV lookup and the dial; it does not affect the established connection.
// Returns the resolved host and port.
func (c *TCPClient) Connect(ctx context.Context, address string) (host string, port string, err error) {
	resolvedAddr := resolveMinecraftAddress(ctx, address)

	var dialer net.Dialer
	netConn, err := dialer.DialContext(ctx, "tcp", resolvedAddr)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to %s: %w", resolvedAddr, err)
	}

	c.conn = NewConn(netConn)
	return net.SplitHostPort(resolvedAddr)
}

// SetConn sets the underlying connection (for testing or server-accepted connections).
func (c *TCPClient) SetConn(conn *Conn) {
	c.conn = conn
}

// Close closes the connection.
func (c *TCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Conn returns the underlying Conn.
func (c *TCPClient) Conn() *Conn {
	return c.conn
}

// SetState sets the protocol state.
func (c *TCPClient) SetState(state State) {
	c.state.Store(uint32(state))
}

// State returns the current protocol state.
func (c *TCPClient) State() State {
	return State(c.state.Load())
}

// SetCompressionThreshold sets the connection's compression threshold.
// Use -1 to disable compression.
func (c *TCPClient) SetCompressionThreshold(threshold int) {
	if c.conn != nil {
		c.conn.SetCompressionThreshold(threshold)
	}
}

// EnableDebug enables or disables debug logging to stdout.
func (c *TCPClient) EnableDebug(enabled bool) {
	if enabled {
		c.logger = log.New(os.Stdout, "[TCPClient] ", log.LstdFlags)
	} else {
		c.logger = nil
	}
}

// WritePacket writes a typed Packet to the connection.
// Safe for concurrent use from multiple goroutines.
func (c *TCPClient) WritePacket(p Packet) error {
	if c.conn == nil {
		return errNilConn
	}
	c.debugf("-> send: state=%v bound=%v id=0x%02X", p.State(), p.Bound(), int(p.ID()))
	return c.conn.WritePacket(p)
}

// WriteWirePacket writes a raw WirePacket to the connection.
// Safe for concurrent use from multiple goroutines.
func (c *TCPClient) WriteWirePacket(pkt *WirePacket) error {
	if c.conn == nil {
		return errNilConn
	}
	c.debugf("-> send (wire): id=0x%02X data_len=%d", int(pkt.PacketID), len(pkt.Data))
	return c.conn.WriteWirePacket(pkt)
}

// ReadWirePacket reads a raw WirePacket from the connection.
func (c *TCPClient) ReadWirePacket() (*WirePacket, error) {
	if c.conn == nil {
		return nil, errNilConn
	}

	c.debugf("<- recv: waiting for packet")
	wire, err := c.conn.ReadWirePacket()
	if err != nil {
		return nil, err
	}

	c.debugf("<- recv: id=0x%02X data_len=%d", int(wire.PacketID), len(wire.Data))
	return wire, nil
}

func (c *TCPClient) debugf(format string, args ...any) {
	if c.logger != nil {
		c.logger.Printf(format, args...)
	}
}

// resolveMinecraftAddress resolves a Minecraft server address using SRV records if available,
// falling back to the default port 25565 if no port is specified.
func resolveMinecraftAddress(ctx context.Context, address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = ""
	}

	// if port is explicitly specified, use it directly
	if port != "" {
		return net.JoinHostPort(host, port)
	}

	// ...otherwise, lookup SRV _minecraft._tcp.<host>
	_, srvRecords, err := net.DefaultResolver.LookupSRV(ctx, "minecraft", "tcp", host)
	if err == nil && len(srvRecords) > 0 {
		srv := srvRecords[0]
		target := strings.TrimSuffix(srv.Target, ".")
		return net.JoinHostPort(target, strconv.Itoa(int(srv.Port)))
	}

	// no SRV record found (or the lookup was canceled: the dial then fails), use default port
	return net.JoinHostPort(host, "25565")
}
