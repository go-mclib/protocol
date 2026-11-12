package wire_test

import (
	"testing"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// Wire counts and lengths come from untrusted peers: negative or oversized
// values must fail cleanly instead of panicking or allocating gigabytes.

var (
	varIntNeg1   = []byte{0xff, 0xff, 0xff, 0xff, 0x0f}
	varIntMax    = []byte{0xff, 0xff, 0xff, 0xff, 0x07}
	varIntMinI32 = []byte{0x80, 0x80, 0x80, 0x80, 0x08}
)

func cat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func noComponents(*ns.PacketBuffer, ns.VarInt) ([]byte, error) { return nil, nil }

func TestMalformedCountsRejected(t *testing.T) {
	tests := []struct {
		name   string
		raw    []byte
		decode func(*ns.PacketBuffer) error
	}{
		{"hashed slot negative add", cat([]byte{1, 1, 1}, varIntNeg1), func(b *ns.PacketBuffer) error {
			_, err := b.ReadHashedSlot()
			return err
		}},
		{"hashed slot huge add", cat([]byte{1, 1, 1}, varIntMax), func(b *ns.PacketBuffer) error {
			_, err := b.ReadHashedSlot()
			return err
		}},
		{"hashed slot negative remove", cat([]byte{1, 1, 1, 0}, varIntNeg1), func(b *ns.PacketBuffer) error {
			_, err := b.ReadHashedSlot()
			return err
		}},
		{"slot negative add", cat([]byte{1, 1}, varIntNeg1, []byte{0}), func(b *ns.PacketBuffer) error {
			_, err := b.ReadSlot(noComponents)
			return err
		}},
		{"slot huge remove", cat([]byte{1, 1, 0}, varIntMax), func(b *ns.PacketBuffer) error {
			_, err := b.ReadSlot(noComponents)
			return err
		}},
		{"fixed byte array negative", nil, func(b *ns.PacketBuffer) error {
			_, err := b.ReadFixedByteArray(-1)
			return err
		}},
		{"fixed byte array huge", []byte{1, 2, 3}, func(b *ns.PacketBuffer) error {
			_, err := b.ReadFixedByteArray(1<<31 - 1)
			return err
		}},
		{"byte array negative", varIntNeg1, func(b *ns.PacketBuffer) error {
			_, err := b.ReadByteArray(0)
			return err
		}},
		{"byte array huge", varIntMax, func(b *ns.PacketBuffer) error {
			_, err := b.ReadByteArray(0)
			return err
		}},
		{"string huge", varIntMax, func(b *ns.PacketBuffer) error {
			_, err := b.ReadString(0)
			return err
		}},
		{"id set negative type", varIntNeg1, func(b *ns.PacketBuffer) error {
			var s ns.IDSet
			return s.Decode(b)
		}},
		{"id set min int32 type", varIntMinI32, func(b *ns.PacketBuffer) error {
			var s ns.IDSet
			return s.Decode(b)
		}},
		{"id set huge type", varIntMax, func(b *ns.PacketBuffer) error {
			var s ns.IDSet
			return s.Decode(b)
		}},
		{"chunk negative heightmap count", varIntNeg1, func(b *ns.PacketBuffer) error {
			var c ns.ChunkData
			return c.Decode(b)
		}},
		{"chunk huge heightmap count", varIntMax, func(b *ns.PacketBuffer) error {
			var c ns.ChunkData
			return c.Decode(b)
		}},
		{"chunk negative heightmap length", cat([]byte{1, 1}, varIntNeg1), func(b *ns.PacketBuffer) error {
			var c ns.ChunkData
			return c.Decode(b)
		}},
		{"chunk negative block entity count", cat([]byte{0, 0}, varIntNeg1), func(b *ns.PacketBuffer) error {
			var c ns.ChunkData
			return c.Decode(b)
		}},
		{"light negative sky count", cat([]byte{0, 0, 0, 0}, varIntNeg1), func(b *ns.PacketBuffer) error {
			var l ns.LightData
			return l.Decode(b)
		}},
		{"light huge block count", cat([]byte{0, 0, 0, 0, 0}, varIntMax), func(b *ns.PacketBuffer) error {
			var l ns.LightData
			return l.Decode(b)
		}},
		{"prefixed array huge", varIntMax, func(b *ns.PacketBuffer) error {
			var a ns.PrefixedArray[ns.VarInt]
			return a.DecodeWith(b, (*ns.PacketBuffer).ReadVarInt)
		}},
		{"bitset huge", varIntMax, func(b *ns.PacketBuffer) error {
			var s ns.BitSet
			return s.Decode(b)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.decode(ns.NewReader(tc.raw)); err == nil {
				t.Fatal("expected error for malformed input")
			}
		})
	}
}
