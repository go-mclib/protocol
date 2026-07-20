package net_structures

import (
	"fmt"

	"github.com/go-mclib/nbt"
)

// NumberFormatType is the registry id from minecraft:number_format_type.
// The ids are stable for the currently supported Minecraft protocol.
type NumberFormatType VarInt

const (
	NumberFormatBlank  NumberFormatType = 0
	NumberFormatStyled NumberFormatType = 1
	NumberFormatFixed  NumberFormatType = 2
)

// NumberFormat is the scoreboard number-format union used by objective and
// score packets. Blank has no payload, styled carries an inline NBT Style, and
// fixed carries a text component.
type NumberFormat struct {
	Type  NumberFormatType
	Style nbt.Compound
	Fixed TextComponent
}

// Decode reads a registry-dispatched scoreboard number format.
func (f *NumberFormat) Decode(buf *PacketBuffer) error {
	kind, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	f.Type = NumberFormatType(kind)
	switch f.Type {
	case NumberFormatBlank:
		return nil
	case NumberFormatStyled:
		r := nbt.NewReaderFrom(buf.Reader())
		tag, _, err := r.ReadTag(true)
		if err != nil {
			return err
		}
		style, ok := tag.(nbt.Compound)
		if !ok {
			return fmt.Errorf("scoreboard styled number format is %T, want NBT compound", tag)
		}
		f.Style = style
		return nil
	case NumberFormatFixed:
		f.Fixed, err = buf.ReadTextComponent()
		return err
	default:
		return fmt.Errorf("unknown scoreboard number format type %d", kind)
	}
}

// Encode writes a registry-dispatched scoreboard number format.
func (f NumberFormat) Encode(buf *PacketBuffer) error {
	if err := buf.WriteVarInt(VarInt(f.Type)); err != nil {
		return err
	}
	switch f.Type {
	case NumberFormatBlank:
		return nil
	case NumberFormatStyled:
		style := f.Style
		if style == nil {
			style = nbt.Compound{}
		}
		data, err := nbt.EncodeNetwork(style)
		if err != nil {
			return err
		}
		_, err = buf.Write(data)
		return err
	case NumberFormatFixed:
		return buf.WriteTextComponent(f.Fixed)
	default:
		return fmt.Errorf("unknown scoreboard number format type %d", f.Type)
	}
}
