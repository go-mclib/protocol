package net_structures

import (
	"encoding/json"
	"errors"
)

// UTF-8 string prefixed with its size in bytes as a VarInt.
// Maximum length of n characters, which varies by context.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:String
type String string

func (s String) ToBytes() (ByteArray, error) {
	bytes := []byte(s)
	lengthBytes, err := VarInt(len(bytes)).ToBytes()
	if err != nil {
		return nil, err
	}
	return append(lengthBytes, bytes...), nil
}

func (s *String) FromBytes(data ByteArray) (int, error) {
	var length VarInt
	bytesRead, err := length.FromBytes(data)
	if err != nil {
		return 0, err
	}

	if int(length) < 0 {
		return 0, errors.New("negative string length")
	}

	if len(data) < bytesRead+int(length) {
		return 0, errors.New("insufficient data for string")
	}

	str := string(data[bytesRead : bytesRead+int(length)])
	*s = String(str)
	return bytesRead + int(length), nil
}

// JSON Text Component - text component serialized as JSON string
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:JSON_Text_Component
type JSONTextComponent map[string]any

func (c JSONTextComponent) ToBytes() (ByteArray, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return String(jsonBytes).ToBytes()
}

func (c *JSONTextComponent) FromBytes(data ByteArray) (int, error) {
	var str String
	bytesRead, err := str.FromBytes(data)
	if err != nil {
		return 0, err
	}

	var component JSONTextComponent
	if err := json.Unmarshal([]byte(str), &component); err != nil {
		return 0, err
	}

	*c = component
	return bytesRead, nil
}

// Identifier - namespaced location (e.g. "minecraft:stone")
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Type:Identifier
type Identifier string

func (id Identifier) ToBytes() (ByteArray, error) {
	return String(id).ToBytes()
}

func (id *Identifier) FromBytes(data ByteArray) (int, error) {
	var str String
	bytesRead, err := str.FromBytes(data)
	if err != nil {
		return 0, err
	}
	*id = Identifier(str)
	return bytesRead, nil
}
