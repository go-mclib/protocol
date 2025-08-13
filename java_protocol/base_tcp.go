package java_protocol

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"

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
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
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
