package javaproto_test

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"net"
	"sync"
	"testing"

	"github.com/go-mclib/protocol/crypto"
	jp "github.com/go-mclib/protocol/javaproto"
	ns "github.com/go-mclib/protocol/javaproto/wire"
)

func zlibBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// frameOf prefixes body with its VarInt length.
func frameOf(body []byte) []byte {
	return append(ns.VarInt(len(body)).Append(nil), body...)
}

func TestReadWirePacketRejectsMalformedFrames(t *testing.T) {
	bomb := zlibBytes(t, make([]byte, 1<<20))
	tests := []struct {
		name      string
		raw       []byte
		threshold int
	}{
		{"negative frame length", []byte{0xff, 0xff, 0xff, 0xff, 0x0f}, -1},
		{"oversized frame length", []byte{0xff, 0xff, 0xff, 0xff, 0x07}, -1},
		{"frame length above 2^21-1", ns.VarInt(1 << 21).Append(nil), -1},
		{"negative data length", frameOf([]byte{0xff, 0xff, 0xff, 0xff, 0x0f, 0}), 0},
		{"data length above 8 MiB", frameOf(append(ns.VarInt(1<<23+1).Append(nil), bomb...)), 0},
		{"zlib bomb exceeds declared length", frameOf(append(ns.VarInt(16).Append(nil), bomb...)), 0},
		{"decompressed shorter than declared", frameOf(append(ns.VarInt(64).Append(nil), zlibBytes(t, []byte{1, 2})...)), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := jp.ReadWirePacketFrom(bytes.NewReader(tc.raw), tc.threshold); err == nil {
				t.Fatal("expected error for malformed frame")
			}
		})
	}
}

func TestWirePacketRoundTrip(t *testing.T) {
	small := &jp.WirePacket{PacketID: 0x2a, Data: []byte("hello")}
	large := &jp.WirePacket{PacketID: 300, Data: bytes.Repeat([]byte("minecraft"), 100)}
	for _, threshold := range []int{-1, 0, 256} {
		for _, want := range []*jp.WirePacket{small, large} {
			var buf bytes.Buffer
			if err := want.WriteTo(&buf, threshold); err != nil {
				t.Fatal(err)
			}
			got, err := jp.ReadWirePacketFrom(&buf, threshold)
			if err != nil {
				t.Fatalf("threshold %d: %v", threshold, err)
			}
			if got.PacketID != want.PacketID || !bytes.Equal(got.Data, want.Data) || buf.Len() != 0 {
				t.Fatalf("threshold %d: round trip mismatch: got id %d, %d data bytes", threshold, got.PacketID, len(got.Data))
			}
		}
	}
}

func TestConnEncryptionSwitchWithBufferedBytes(t *testing.T) {
	secret := bytes.Repeat([]byte{7}, 16)
	plain := &jp.WirePacket{PacketID: 1, Data: []byte("before")}
	secure := &jp.WirePacket{PacketID: 2, Data: []byte("after")}

	var first, second bytes.Buffer
	if err := plain.WriteTo(&first, -1); err != nil {
		t.Fatal(err)
	}
	if err := secure.WriteTo(&second, -1); err != nil {
		t.Fatal(err)
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		t.Fatal(err)
	}
	encrypted := second.Bytes()
	crypto.NewCFB8(block, secret, false).XORKeyStream(encrypted, encrypted)

	local, remote := net.Pipe()
	defer local.Close()
	defer remote.Close()
	go func() {
		// one write, so the reader buffers both frames before encryption is enabled
		_, _ = remote.Write(append(first.Bytes(), encrypted...))
	}()

	conn := jp.NewConn(local)
	got, err := conn.ReadWirePacket()
	if err != nil || string(got.Data) != "before" {
		t.Fatalf("plain packet: %v %+v", err, got)
	}
	if err := conn.EnableEncryption(secret); err != nil {
		t.Fatal(err)
	}
	got, err = conn.ReadWirePacket()
	if err != nil || string(got.Data) != "after" {
		t.Fatalf("encrypted packet: %v %+v", err, got)
	}
}

func TestConnConcurrentEncryptedWrites(t *testing.T) {
	secret := bytes.Repeat([]byte{3}, 16)
	local, remote := net.Pipe()
	defer local.Close()
	defer remote.Close()
	writer, reader := jp.NewConn(local), jp.NewConn(remote)
	for _, c := range []*jp.Conn{writer, reader} {
		c.SetCompressionThreshold(8)
		if err := c.EnableEncryption(secret); err != nil {
			t.Fatal(err)
		}
	}

	const n = 50
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			pkt := &jp.WirePacket{PacketID: ns.VarInt(i), Data: bytes.Repeat([]byte{byte(i)}, i)}
			if err := writer.WriteWirePacket(pkt); err != nil {
				t.Error(err)
			}
		})
	}
	seen := make(map[ns.VarInt]bool)
	for range n {
		got, err := reader.ReadWirePacket()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.Data, bytes.Repeat([]byte{byte(got.PacketID)}, int(got.PacketID))) {
			t.Fatalf("packet %d corrupted", got.PacketID)
		}
		seen[got.PacketID] = true
	}
	wg.Wait()
	if len(seen) != n {
		t.Fatalf("received %d distinct packets, want %d", len(seen), n)
	}
}

func TestConnEnableEncryptionRejectsNon16ByteSecret(t *testing.T) {
	local, remote := net.Pipe()
	defer local.Close()
	defer remote.Close()
	conn := jp.NewConn(local)
	for _, n := range []int{0, 15, 24, 32} {
		if err := conn.EnableEncryption(make([]byte, n)); err == nil {
			t.Errorf("accepted %d-byte shared secret", n)
		}
	}
	if conn.Encrypted() {
		t.Fatal("rejected secret enabled encryption")
	}
}
