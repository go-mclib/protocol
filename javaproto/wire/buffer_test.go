package wire_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

func TestBufferReaderModes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}

	t.Run("NewReader from bytes", func(t *testing.T) {
		buf := ns.NewReader(data)
		b, err := buf.ReadByte()
		if err != nil {
			t.Fatalf("ReadByte() error = %v", err)
		}
		if b != 0x01 {
			t.Errorf("ReadByte() = %v, want %v", b, 0x01)
		}
	})
}

func TestBufferWriterModes(t *testing.T) {
	t.Run("NewWriter to internal buffer", func(t *testing.T) {
		buf := ns.NewWriter()
		if err := buf.WriteByte(0x42); err != nil {
			t.Fatalf("WriteByte() error = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0x42}) {
			t.Errorf("Bytes() = %v, want %v", buf.Bytes(), []byte{0x42})
		}
	})
}

func TestBufferReadExact(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	buf := ns.NewReader(data)

	// read exactly 2 bytes
	p := make([]byte, 2)
	n, err := buf.Read(p)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != 2 {
		t.Errorf("Read() n = %d, want 2", n)
	}
	if !bytes.Equal(p, []byte{0x01, 0x02}) {
		t.Errorf("Read() = %v, want %v", p, []byte{0x01, 0x02})
	}

	// read remaining
	p = make([]byte, 1)
	n, err = buf.Read(p)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != 1 || p[0] != 0x03 {
		t.Errorf("Read() = %v, want %v", p, []byte{0x03})
	}
}

func TestBufferReadEOF(t *testing.T) {
	data := []byte{0x01}
	buf := ns.NewReader(data)

	// read 1 byte, should succeed
	_, err := buf.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte() error = %v", err)
	}

	// read another, should fail
	_, err = buf.ReadByte()
	if err != io.EOF {
		t.Errorf("ReadByte() error = %v, want io.EOF", err)
	}
}

func TestBufferByteArrayReadWrite(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// write
	buf := ns.NewWriter()
	if err := buf.WriteByteArray(data); err != nil {
		t.Fatalf("WriteByteArray() error = %v", err)
	}

	// should be: VarInt(5) + data = 0x05 + data
	expected := append([]byte{0x05}, data...)
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("WriteByteArray() = %v, want %v", buf.Bytes(), expected)
	}

	// read
	reader := ns.NewReader(buf.Bytes())
	got, err := reader.ReadByteArray(0)
	if err != nil {
		t.Fatalf("ReadByteArray() error = %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("ReadByteArray() = %v, want %v", got, data)
	}
}

func TestBufferByteArrayMaxLen(t *testing.T) {
	// create byte array with 10 bytes
	data := make([]byte, 10)
	buf := ns.NewWriter()
	_ = buf.WriteByteArray(data)

	// try to read with max 5 bytes - should fail
	reader := ns.NewReader(buf.Bytes())
	_, err := reader.ReadByteArray(5)
	if err == nil {
		t.Error("ReadByteArray() should error when exceeding max length")
	}
}

func TestBufferFixedByteArray(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}

	// write
	buf := ns.NewWriter()
	if err := buf.WriteFixedByteArray(data); err != nil {
		t.Fatalf("WriteFixedByteArray() error = %v", err)
	}

	// should be raw bytes, no length prefix
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("WriteFixedByteArray() = %v, want %v", buf.Bytes(), data)
	}

	// read
	reader := ns.NewReader(buf.Bytes())
	got, err := reader.ReadFixedByteArray(3)
	if err != nil {
		t.Fatalf("ReadFixedByteArray() error = %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("ReadFixedByteArray() = %v, want %v", got, data)
	}
}

func TestBufferReset(t *testing.T) {
	buf := ns.NewWriter()
	_ = buf.WriteByte(0x01)
	_ = buf.WriteByte(0x02)

	if buf.Len() != 2 {
		t.Errorf("Len() = %d, want 2", buf.Len())
	}

	buf.Reset()

	if buf.Len() != 0 {
		t.Errorf("Len() after Reset() = %d, want 0", buf.Len())
	}
	if len(buf.Bytes()) != 0 {
		t.Errorf("Bytes() after Reset() = %v, want empty", buf.Bytes())
	}
}

func TestBufferModeErrors(t *testing.T) {
	t.Run("write on reader", func(t *testing.T) {
		buf := ns.NewReader([]byte{0x01})
		_, err := buf.Write([]byte{0x02})
		if err == nil {
			t.Error("Write() on reader should error")
		}
	})

	t.Run("read on writer", func(t *testing.T) {
		buf := ns.NewWriter()
		_, err := buf.Read(make([]byte, 1))
		if err == nil {
			t.Error("Read() on writer should error")
		}
	})

	// typed accessors must report the mode error instead of panicking on the nil side
	t.Run("typed read on writer", func(t *testing.T) {
		buf := ns.NewWriter()
		reads := map[string]func() error{
			"VarInt":     func() error { _, err := buf.ReadVarInt(); return err },
			"VarLong":    func() error { _, err := buf.ReadVarLong(); return err },
			"Bool":       func() error { _, err := buf.ReadBool(); return err },
			"Int8":       func() error { _, err := buf.ReadInt8(); return err },
			"Uint8":      func() error { _, err := buf.ReadUint8(); return err },
			"Int16":      func() error { _, err := buf.ReadInt16(); return err },
			"Uint16":     func() error { _, err := buf.ReadUint16(); return err },
			"Int32":      func() error { _, err := buf.ReadInt32(); return err },
			"Int64":      func() error { _, err := buf.ReadInt64(); return err },
			"Float32":    func() error { _, err := buf.ReadFloat32(); return err },
			"Float64":    func() error { _, err := buf.ReadFloat64(); return err },
			"String":     func() error { _, err := buf.ReadString(0); return err },
			"Identifier": func() error { _, err := buf.ReadIdentifier(); return err },
			"Position":   func() error { _, err := buf.ReadPosition(); return err },
			"GlobalPos":  func() error { _, err := buf.ReadGlobalPos(); return err },
			"UUID":       func() error { _, err := buf.ReadUUID(); return err },
			"Angle":      func() error { _, err := buf.ReadAngle(); return err },
		}
		for name, read := range reads {
			if err := read(); err == nil || !strings.Contains(err.Error(), "not in read mode") {
				t.Errorf("Read%s() on writer = %v, want read-mode error", name, err)
			}
		}
	})

	t.Run("typed write on reader", func(t *testing.T) {
		buf := ns.NewReader(nil)
		writes := map[string]error{
			"VarInt":     buf.WriteVarInt(1),
			"VarLong":    buf.WriteVarLong(1),
			"Bool":       buf.WriteBool(true),
			"Int8":       buf.WriteInt8(1),
			"Uint8":      buf.WriteUint8(1),
			"Int16":      buf.WriteInt16(1),
			"Uint16":     buf.WriteUint16(1),
			"Int32":      buf.WriteInt32(1),
			"Int64":      buf.WriteInt64(1),
			"Float32":    buf.WriteFloat32(1),
			"Float64":    buf.WriteFloat64(1),
			"String":     buf.WriteString("a"),
			"Identifier": buf.WriteIdentifier("a:b"),
			"Position":   buf.WritePosition(ns.Position{}),
			"GlobalPos":  buf.WriteGlobalPos(ns.GlobalPos{}),
			"UUID":       buf.WriteUUID(ns.UUID{}),
			"Angle":      buf.WriteAngle(1),
		}
		for name, err := range writes {
			if err == nil || !strings.Contains(err.Error(), "not in write mode") {
				t.Errorf("Write%s() on reader = %v, want write-mode error", name, err)
			}
		}
	})
}
