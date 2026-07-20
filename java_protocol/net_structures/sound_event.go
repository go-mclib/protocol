package net_structures

import "fmt"

// SoundEvent is a Holder<SoundEvent>. Registry references encode as the
// registry id plus one; direct holders encode zero followed by the sound id and
// optional fixed range.
type SoundEvent struct {
	Direct     bool
	RegistryID VarInt
	SoundID    Identifier
	FixedRange PrefixedOptional[Float32]
}

// RegistrySoundEvent constructs a registry-backed sound holder.
func RegistrySoundEvent(id int32) SoundEvent {
	return SoundEvent{RegistryID: VarInt(id)}
}

// DirectSoundEvent constructs a direct variable-range sound holder.
func DirectSoundEvent(id Identifier) SoundEvent {
	return SoundEvent{Direct: true, SoundID: id}
}

func (s *SoundEvent) Decode(buf *PacketBuffer) error {
	marker, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	if marker != 0 {
		s.Direct = false
		s.RegistryID = marker - 1
		return nil
	}
	s.Direct = true
	if s.SoundID, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	return s.FixedRange.DecodeWith(buf, func(b *PacketBuffer) (Float32, error) { return b.ReadFloat32() })
}

func (s SoundEvent) Encode(buf *PacketBuffer) error {
	if !s.Direct {
		if s.RegistryID < 0 {
			return fmt.Errorf("invalid sound event registry id %d", s.RegistryID)
		}
		return buf.WriteVarInt(s.RegistryID + 1)
	}
	if s.SoundID == "" {
		return fmt.Errorf("direct sound event has empty id")
	}
	if err := buf.WriteVarInt(0); err != nil {
		return err
	}
	if err := buf.WriteIdentifier(s.SoundID); err != nil {
		return err
	}
	return s.FixedRange.EncodeWith(buf, func(b *PacketBuffer, v Float32) error { return b.WriteFloat32(v) })
}
