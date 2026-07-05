package net_structures

import (
	"fmt"
)

// ElementEncoder is a function that encodes an element to a buffer.
type ElementEncoder[T any] func(buf *PacketBuffer, v T) error

// ElementDecoder is a function that decodes an element from a buffer.
type ElementDecoder[T any] func(buf *PacketBuffer) (T, error)

// PrefixedArray is a VarInt length-prefixed array of elements.
//
// Wire format:
//
//	┌───────────────────────┬───────────────────────────────┐
//	│  Length (VarInt)      │  Elements (T × Length)        │
//	└───────────────────────┴───────────────────────────────┘
//
// Example usage:
//
//	type MyPacket struct {
//	    Names PrefixedArray[String]
//	}
//
//	// in Read:
//	p.Names.DecodeWith(buf, func(b *PacketBuffer) (String, error) {
//	    return b.ReadString(32767)
//	})
//
//	// in Write:
//	p.Names.EncodeWith(buf, func(b *PacketBuffer, v String) error {
//	    return b.WriteString(v)
//	})
type PrefixedArray[T any] []T

// DecodeWith reads a length-prefixed array using the provided decoder function.
func (a *PrefixedArray[T]) DecodeWith(buf *PacketBuffer, decode ElementDecoder[T]) error {
	length, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read array length: %w", err)
	}
	if length < 0 {
		return fmt.Errorf("negative array length: %d", length)
	}

	*a = make([]T, length)
	for i := range *a {
		(*a)[i], err = decode(buf)
		if err != nil {
			return fmt.Errorf("failed to read array element %d: %w", i, err)
		}
	}
	return nil
}

// EncodeWith writes a length-prefixed array using the provided encoder function.
func (a PrefixedArray[T]) EncodeWith(buf *PacketBuffer, encode ElementEncoder[T]) error {
	if err := buf.WriteVarInt(VarInt(len(a))); err != nil {
		return fmt.Errorf("failed to write array length: %w", err)
	}
	for i, v := range a {
		if err := encode(buf, v); err != nil {
			return fmt.Errorf("failed to write array element %d: %w", i, err)
		}
	}
	return nil
}

// Len returns the number of elements in the array.
func (a PrefixedArray[T]) Len() int {
	return len(a)
}

// PrefixedOptional is a Boolean-prefixed optional value.
//
// Wire format:
//
//	┌──────────────────┬─────────────────────────────────────┐
//	│  Present (Bool)  │  Value (T, only if Present=true)    │
//	└──────────────────┴─────────────────────────────────────┘
//
// Example usage:
//
//	type MyPacket struct {
//	    Title PrefixedOptional[String]
//	}
//
//	// In Read:
//	p.Title.DecodeWith(buf, func(b *PacketBuffer) (String, error) {
//	    return b.ReadString(32767)
//	})
//
//	// In Write:
//	p.Title.EncodeWith(buf, func(b *PacketBuffer, v String) error {
//	    return b.WriteString(v)
//	})
type PrefixedOptional[T any] struct {
	Present bool
	Value   T
}

// Some creates a PrefixedOptional with a value.
func Some[T any](value T) PrefixedOptional[T] {
	return PrefixedOptional[T]{Present: true, Value: value}
}

// None creates an empty PrefixedOptional.
func None[T any]() PrefixedOptional[T] {
	return PrefixedOptional[T]{Present: false}
}

// DecodeWith reads a boolean-prefixed optional using the provided decoder.
func (o *PrefixedOptional[T]) DecodeWith(buf *PacketBuffer, decode ElementDecoder[T]) error {
	present, err := buf.ReadBool()
	if err != nil {
		return fmt.Errorf("failed to read optional presence: %w", err)
	}
	o.Present = bool(present)

	if o.Present {
		o.Value, err = decode(buf)
		if err != nil {
			return fmt.Errorf("failed to read optional value: %w", err)
		}
	}
	return nil
}

// EncodeWith writes a boolean-prefixed optional using the provided encoder.
func (o PrefixedOptional[T]) EncodeWith(buf *PacketBuffer, encode ElementEncoder[T]) error {
	if err := buf.WriteBool(Boolean(o.Present)); err != nil {
		return fmt.Errorf("failed to write optional presence: %w", err)
	}
	if o.Present {
		if err := encode(buf, o.Value); err != nil {
			return fmt.Errorf("failed to write optional value: %w", err)
		}
	}
	return nil
}

// Get returns the value and whether it's present.
func (o PrefixedOptional[T]) Get() (T, bool) {
	return o.Value, o.Present
}
