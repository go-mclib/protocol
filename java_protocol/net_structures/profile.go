package net_structures

import (
	"fmt"
)

// ProfileProperty represents a single property in a game profile.
type ProfileProperty struct {
	Name      String
	Value     String
	Signature PrefixedOptional[String]
}

// Decode reads a ProfileProperty from the buffer.
func (p *ProfileProperty) Decode(buf *PacketBuffer) error {
	var err error
	p.Name, err = buf.ReadString(64)
	if err != nil {
		return fmt.Errorf("failed to read property name: %w", err)
	}
	p.Value, err = buf.ReadString(32767)
	if err != nil {
		return fmt.Errorf("failed to read property value: %w", err)
	}
	return decodeOptional(buf, &p.Signature, "property signature", stringDecoder(1024))
}

// Encode writes a ProfileProperty to the buffer.
func (p *ProfileProperty) Encode(buf *PacketBuffer) error {
	if err := buf.WriteString(p.Name); err != nil {
		return fmt.Errorf("failed to write property name: %w", err)
	}
	if err := buf.WriteString(p.Value); err != nil {
		return fmt.Errorf("failed to write property value: %w", err)
	}
	return encodeOptional(buf, p.Signature, "property signature", (*PacketBuffer).WriteString)
}

// GameProfile represents a complete player profile with UUID, username, and properties.
//
// Wire format:
//
//	┌───────────────────┬─────────────────────────────────────────────────┐
//	│  UUID (16 bytes)  │  Username (String, max 16)                      │
//	├───────────────────┼─────────────────────────────────────────────────┤
//	│  Properties (VarInt length + array of ProfileProperty)              │
//	└─────────────────────────────────────────────────────────────────────┘
type GameProfile struct {
	UUID       UUID
	Username   String
	Properties PrefixedArray[ProfileProperty]
}

// Decode reads a GameProfile from the buffer.
func (p *GameProfile) Decode(buf *PacketBuffer) error {
	var err error
	p.UUID, err = buf.ReadUUID()
	if err != nil {
		return fmt.Errorf("failed to read profile uuid: %w", err)
	}
	p.Username, err = buf.ReadString(16)
	if err != nil {
		return fmt.Errorf("failed to read profile username: %w", err)
	}
	if err := p.Properties.DecodeWith(buf, decodeProfileProperty); err != nil {
		return fmt.Errorf("failed to read profile properties: %w", err)
	}
	return nil
}

// Encode writes a GameProfile to the buffer.
func (p *GameProfile) Encode(buf *PacketBuffer) error {
	if err := buf.WriteUUID(p.UUID); err != nil {
		return fmt.Errorf("failed to write profile uuid: %w", err)
	}
	if err := buf.WriteString(p.Username); err != nil {
		return fmt.Errorf("failed to write profile username: %w", err)
	}
	if err := p.Properties.EncodeWith(buf, encodeProfileProperty); err != nil {
		return fmt.Errorf("failed to write profile properties: %w", err)
	}
	return nil
}

// ReadGameProfile reads a GameProfile from the buffer.
func (pb *PacketBuffer) ReadGameProfile() (GameProfile, error) {
	var p GameProfile
	err := p.Decode(pb)
	return p, err
}

// WriteGameProfile writes a GameProfile to the buffer.
func (pb *PacketBuffer) WriteGameProfile(p GameProfile) error {
	return p.Encode(pb)
}

// ResolvableProfileKind indicates whether a profile is partial or complete.
type ResolvableProfileKind VarInt

const (
	ProfilePartial  ResolvableProfileKind = 0
	ProfileComplete ResolvableProfileKind = 1
)

// ResolvableProfile represents a player profile that can be either partial or complete.
//
// Wire format:
//
//	┌─────────────────────┬───────────────────────────────────────────────┐
//	│  Kind (VarInt Enum) │  Data (depends on Kind)                       │
//	└─────────────────────┴───────────────────────────────────────────────┘
//
// Partial (0): Optional username, optional UUID, optional properties, optional signature
// Complete (1): Full GameProfile + optional body/cape/elytra/model identifiers
type ResolvableProfile struct {
	Kind ResolvableProfileKind

	// partial profile fields (Kind = 0)
	PartialUsername   PrefixedOptional[String]
	PartialUUID       PrefixedOptional[UUID]
	PartialProperties PrefixedOptional[PrefixedArray[ProfileProperty]]
	PartialSignature  PrefixedOptional[String]

	// complete profile fields (Kind = 1)
	CompleteProfile GameProfile
	BodyModel       PrefixedOptional[Identifier]
	CapeModel       PrefixedOptional[Identifier]
	ElytraModel     PrefixedOptional[Identifier]
	SkinModel       PrefixedOptional[VarInt] // enum: 0=wide, 1=slim
}

// NewPartialProfile creates a partial resolvable profile.
func NewPartialProfile() *ResolvableProfile {
	return &ResolvableProfile{Kind: ProfilePartial}
}

// NewCompleteProfile creates a complete resolvable profile from a game profile.
func NewCompleteProfile(profile GameProfile) *ResolvableProfile {
	return &ResolvableProfile{
		Kind:            ProfileComplete,
		CompleteProfile: profile,
	}
}

// Decode reads a ResolvableProfile from the buffer.
func (p *ResolvableProfile) Decode(buf *PacketBuffer) error {
	kind, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read resolvable profile kind: %w", err)
	}
	p.Kind = ResolvableProfileKind(kind)

	switch p.Kind {
	case ProfilePartial:
		if err := decodeOptional(buf, &p.PartialUsername, "partial username", stringDecoder(16)); err != nil {
			return err
		}
		if err := decodeOptional(buf, &p.PartialUUID, "partial uuid", (*PacketBuffer).ReadUUID); err != nil {
			return err
		}
		if err := decodeOptional(buf, &p.PartialProperties, "partial properties", decodeProfileProperties); err != nil {
			return err
		}
		if err := decodeOptional(buf, &p.PartialSignature, "partial signature", stringDecoder(1024)); err != nil {
			return err
		}

	case ProfileComplete:
		if err := p.CompleteProfile.Decode(buf); err != nil {
			return fmt.Errorf("failed to read complete profile: %w", err)
		}
		if err := decodeOptional(buf, &p.BodyModel, "body model", (*PacketBuffer).ReadIdentifier); err != nil {
			return err
		}
		if err := decodeOptional(buf, &p.CapeModel, "cape model", (*PacketBuffer).ReadIdentifier); err != nil {
			return err
		}
		if err := decodeOptional(buf, &p.ElytraModel, "elytra model", (*PacketBuffer).ReadIdentifier); err != nil {
			return err
		}
		if err := decodeOptional(buf, &p.SkinModel, "skin model", (*PacketBuffer).ReadVarInt); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown resolvable profile kind: %d", p.Kind)
	}
	return nil
}

// Encode writes a ResolvableProfile to the buffer.
func (p *ResolvableProfile) Encode(buf *PacketBuffer) error {
	if err := buf.WriteVarInt(VarInt(p.Kind)); err != nil {
		return fmt.Errorf("failed to write resolvable profile kind: %w", err)
	}

	switch p.Kind {
	case ProfilePartial:
		if err := encodeOptional(buf, p.PartialUsername, "partial username", (*PacketBuffer).WriteString); err != nil {
			return err
		}
		if err := encodeOptional(buf, p.PartialUUID, "partial uuid", (*PacketBuffer).WriteUUID); err != nil {
			return err
		}
		if err := encodeOptional(buf, p.PartialProperties, "partial properties", encodeProfileProperties); err != nil {
			return err
		}
		if err := encodeOptional(buf, p.PartialSignature, "partial signature", (*PacketBuffer).WriteString); err != nil {
			return err
		}

	case ProfileComplete:
		if err := p.CompleteProfile.Encode(buf); err != nil {
			return fmt.Errorf("failed to write complete profile: %w", err)
		}
		if err := encodeOptional(buf, p.BodyModel, "body model", (*PacketBuffer).WriteIdentifier); err != nil {
			return err
		}
		if err := encodeOptional(buf, p.CapeModel, "cape model", (*PacketBuffer).WriteIdentifier); err != nil {
			return err
		}
		if err := encodeOptional(buf, p.ElytraModel, "elytra model", (*PacketBuffer).WriteIdentifier); err != nil {
			return err
		}
		if err := encodeOptional(buf, p.SkinModel, "skin model", (*PacketBuffer).WriteVarInt); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown resolvable profile kind: %d", p.Kind)
	}
	return nil
}

// ReadResolvableProfile reads a ResolvableProfile from the buffer.
func (pb *PacketBuffer) ReadResolvableProfile() (ResolvableProfile, error) {
	var p ResolvableProfile
	err := p.Decode(pb)
	return p, err
}

// WriteResolvableProfile writes a ResolvableProfile to the buffer.
func (pb *PacketBuffer) WriteResolvableProfile(p ResolvableProfile) error {
	return p.Encode(pb)
}

// decodeOptional reads a boolean-prefixed optional field, wrapping errors with the field name.
func decodeOptional[T any](buf *PacketBuffer, o *PrefixedOptional[T], field string, decode ElementDecoder[T]) error {
	if err := o.DecodeWith(buf, decode); err != nil {
		return fmt.Errorf("failed to read %s: %w", field, err)
	}
	return nil
}

// encodeOptional writes a boolean-prefixed optional field, wrapping errors with the field name.
func encodeOptional[T any](buf *PacketBuffer, o PrefixedOptional[T], field string, encode ElementEncoder[T]) error {
	if err := o.EncodeWith(buf, encode); err != nil {
		return fmt.Errorf("failed to write %s: %w", field, err)
	}
	return nil
}

func stringDecoder(maxLen int) ElementDecoder[String] {
	return func(b *PacketBuffer) (String, error) { return b.ReadString(maxLen) }
}

func decodeProfileProperty(b *PacketBuffer) (ProfileProperty, error) {
	var prop ProfileProperty
	err := prop.Decode(b)
	return prop, err
}

func encodeProfileProperty(b *PacketBuffer, prop ProfileProperty) error {
	return prop.Encode(b)
}

func decodeProfileProperties(b *PacketBuffer) (PrefixedArray[ProfileProperty], error) {
	var props PrefixedArray[ProfileProperty]
	err := props.DecodeWith(b, decodeProfileProperty)
	return props, err
}

func encodeProfileProperties(b *PacketBuffer, props PrefixedArray[ProfileProperty]) error {
	return props.EncodeWith(b, encodeProfileProperty)
}
