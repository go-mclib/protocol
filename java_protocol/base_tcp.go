package java_protocol

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	mc_crypto "github.com/go-mclib/protocol/crypto"
)

type BaseTCP struct {
	conn       net.Conn
	encryption *mc_crypto.Encryption
	debug      bool
	logger     *log.Logger
}

func NewBaseTCP(conn net.Conn) *BaseTCP {
	return &BaseTCP{
		conn:       conn,
		encryption: mc_crypto.NewEncryption(),
		debug:      false,
		logger:     log.New(os.Stdout, "[java_protocol]", log.LstdFlags),
	}
}

func (b *BaseTCP) Connect(address string) error {
	resolvedAddr, err := resolveMinecraftAddress(address)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.Dial("tcp", resolvedAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", resolvedAddr, err)
	}
	b.conn = conn
	return nil
}

func (b *BaseTCP) Close() error {
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *BaseTCP) GetConn() net.Conn {
	return b.conn
}

func (b *BaseTCP) GetEncryption() *mc_crypto.Encryption {
	return b.encryption
}

func (b *BaseTCP) EnableDebug(enabled bool) {
	b.debug = enabled
}

func (b *BaseTCP) SetLogger(l *log.Logger) {
	b.logger = l
}

func (b *BaseTCP) logf(format string, args ...any) {
	if b.logger != nil {
		b.logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (b *BaseTCP) debugf(format string, args ...any) {
	if b.debug {
		b.logf(format, args...)
	}
}

// resolveMinecraftAddress resolves a Minecraft server address using SRV records
// if available, falling back to the default port 25565, if no port is specified.
func resolveMinecraftAddress(address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		// no port specified, treat entire address as hostname
		host = address
		port = ""
	}

	// if port is explicitly specified, use it directly without SRV lookup
	if port != "" {
		return net.JoinHostPort(host, port), nil
	}

	// lookup SRV _minecraft._tcp.<host>
	_, srvRecords, err := net.LookupSRV("minecraft", "tcp", host)
	if err == nil && len(srvRecords) > 0 {
		srv := srvRecords[0]
		target := strings.TrimSuffix(srv.Target, ".")
		return net.JoinHostPort(target, strconv.Itoa(int(srv.Port))), nil
	}

	// no SRV record found, use default port
	return net.JoinHostPort(host, "25565"), nil
}

// hexSnippet returns a hex string of at most max bytes of data (for debugging)
func hexSnippet(data []byte, max int) string {
	if data == nil {
		return ""
	}
	if max > 0 && len(data) > max {
		return hex.EncodeToString(data[:max]) + "..."
	}
	return hex.EncodeToString(data)
}
