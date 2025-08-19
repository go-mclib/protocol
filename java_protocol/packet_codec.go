package java_protocol

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	ns "github.com/go-mclib/protocol/net_structures"
)

// parseLengthFromTag extracts length value from mc tag (e.g., "length:256" returns 256)
func parseLengthFromTag(tag string) int {
	if tag == "" {
		return 0
	}

	parts := strings.SplitSeq(tag, ",")
	for part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "length:") {
			lengthStr := strings.TrimPrefix(part, "length:")
			if length, err := strconv.Atoi(lengthStr); err == nil {
				return length
			}
		}
	}
	return 0
}

// PacketDataToBytes converts a struct to bytes using reflection
func PacketDataToBytes(v any) (ns.ByteArray, error) {
	val := reflect.ValueOf(v)

	// handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("cannot marshal nil pointer")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("can only marshal structs, got %v", val.Kind())
	}

	return marshalStruct(val)
}

func marshalStruct(val reflect.Value) (ns.ByteArray, error) {
	var result ns.ByteArray
	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// parse struct tags for configuration
		tag := fieldType.Tag.Get("mc")
		if tag == "-" {
			continue // skip
		}

		// marshal field
		bytes, err := marshalField(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling field %s: %w", fieldType.Name, err)
		}

		result = append(result, bytes...)
	}

	return result, nil
}

func marshalField(field reflect.Value) (ns.ByteArray, error) {
	// has ToBytes method?
	if field.CanAddr() {
		if method := field.Addr().MethodByName("ToBytes"); method.IsValid() {
			results := method.Call(nil)
			if len(results) == 2 {
				if !results[1].IsNil() {
					return nil, results[1].Interface().(error)
				}
				return results[0].Interface().(ns.ByteArray), nil
			}
		}
	}

	// if not addressable, try on the value itself
	if method := field.MethodByName("ToBytes"); method.IsValid() {
		results := method.Call(nil)
		if len(results) == 2 {
			if !results[1].IsNil() {
				return nil, results[1].Interface().(error)
			}
			return results[0].Interface().(ns.ByteArray), nil
		}
	}

	// handle other types
	switch field.Kind() {
	case reflect.Struct:
		// recursively marshal nested structs
		return marshalStruct(field)

	case reflect.Slice:
		// for slices, we need to encode length first
		length := field.Len()
		lengthBytes, err := ns.VarInt(length).ToBytes()
		if err != nil {
			return nil, err
		}

		result := lengthBytes
		for j := range length {
			elemBytes, err := marshalField(field.Index(j))
			if err != nil {
				return nil, err
			}
			result = append(result, elemBytes...)
		}
		return result, nil

	case reflect.Array:
		// fixed-size arrays don't encode length
		var result ns.ByteArray
		for j := range field.Len() {
			elemBytes, err := marshalField(field.Index(j))
			if err != nil {
				return nil, err
			}
			result = append(result, elemBytes...)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported type: %v", field.Type())
	}
}

// BytesToPacketData converts bytes to a struct using reflection
func BytesToPacketData(data ns.ByteArray, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("unmarshal requires a non-nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("can only unmarshal into structs, got %v", elem.Kind())
	}

	_, err := unmarshalStruct(elem, data)
	return err
}

func unmarshalStruct(val reflect.Value, data ns.ByteArray) (int, error) {
	typ := val.Type()
	offset := 0

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// skip unexported
		if !field.CanSet() {
			continue
		}

		// parse tags
		tag := fieldType.Tag.Get("mc")
		if tag == "-" {
			continue // skip
		}

		if offset >= len(data) {
			fieldTypeName := field.Type().String()
			if strings.Contains(fieldTypeName, "PrefixedOptional") {
				presentField := field.FieldByName("Present")
				if presentField.IsValid() && presentField.CanSet() {
					presentField.SetBool(false)
					continue // Skip to next field
				}
			}
			return offset, fmt.Errorf("insufficient data for field %s (have %d bytes left, at offset %d of %d total)", fieldType.Name, len(data)-offset, offset, len(data))
		}

		if i > 0 {
			prevField := val.Field(i - 1)
			prevFieldType := typ.Field(i - 1)
			if prevFieldType.Name == "MessageID" && strings.Contains(field.Type().String(), "Optional") && strings.Contains(field.Type().String(), "FixedByteArray") {
				var prevIsZero bool
				switch prevField.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					prevIsZero = prevField.Int() == 0
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					prevIsZero = prevField.Uint() == 0
				default:
					if vi, ok := prevField.Interface().(ns.VarInt); ok {
						prevIsZero = int(vi) == 0
					}
				}
				if presentField := field.FieldByName("Present"); presentField.IsValid() && presentField.CanSet() {
					presentField.SetBool(prevIsZero)
				}
			}
		}

		bytesConsumed, err := UnmarshalFieldWithTag(field, data[offset:], tag)
		if err != nil {
			return offset, fmt.Errorf("error unmarshaling field %s (at offset %d, %d bytes remaining): %w", fieldType.Name, offset, len(data)-offset, err)
		}

		offset += bytesConsumed
	}

	return offset, nil
}

// unmarshalSliceWithReflection handles slice types (including PrefixedArray) using reflection
// when their FromBytes method fails due to struct elements not implementing FromBytes
func unmarshalSliceWithReflection(field reflect.Value, data ns.ByteArray) (int, error) {
	var length ns.VarInt
	bytesRead, err := length.FromBytes(data)
	if err != nil {
		return 0, err
	}

	if length < 0 {
		return 0, fmt.Errorf("negative array length")
	}

	slice := reflect.MakeSlice(field.Type(), int(length), int(length))
	offset := bytesRead

	for i := 0; i < int(length); i++ {
		elem := slice.Index(i)
		itemBytes, err := UnmarshalField(elem, data[offset:])
		if err != nil {
			return 0, fmt.Errorf("error unmarshaling array item %d: %w", i, err)
		}
		offset += itemBytes
	}

	field.Set(slice)
	return offset, nil
}

// unmarshalPrefixedOptionalFixedByteArray handles PrefixedOptional[FixedByteArray] with length from tag
func unmarshalPrefixedOptionalFixedByteArray(field reflect.Value, data ns.ByteArray, length int) (int, error) {
	var present ns.Boolean
	bytesRead, err := present.FromBytes(data)
	if err != nil {
		return 0, err
	}

	presentField := field.FieldByName("Present")
	if presentField.IsValid() {
		presentField.SetBool(bool(present))
	}

	if !bool(present) {
		return bytesRead, nil
	}

	valueField := field.FieldByName("Value")
	if valueField.IsValid() && valueField.CanSet() {
		fba := ns.FixedByteArray{Length: length}
		valueBytes, err := fba.FromBytes(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		valueField.Set(reflect.ValueOf(fba))
		return bytesRead + valueBytes, nil
	}

	return 0, fmt.Errorf("could not find or set Value field in PrefixedOptional")
}

// unmarshalOptionalFixedByteArray handles Optional[FixedByteArray] with length from tag
func unmarshalOptionalFixedByteArray(field reflect.Value, data ns.ByteArray, length int) (int, error) {
	presentField := field.FieldByName("Present")
	if !presentField.IsValid() || !presentField.Bool() {
		return 0, nil
	}

	valueField := field.FieldByName("Value")
	if valueField.IsValid() && valueField.CanSet() {
		fba := ns.FixedByteArray{Length: length}
		valueBytes, err := fba.FromBytes(data)
		if err != nil {
			return 0, err
		}
		valueField.Set(reflect.ValueOf(fba))
		return valueBytes, nil
	}

	return 0, fmt.Errorf("could not find or set Value field in Optional")
}

// UnmarshalField unmarshals a field without tag information (for backwards compatibility)
func UnmarshalField(field reflect.Value, data ns.ByteArray) (int, error) {
	return UnmarshalFieldWithTag(field, data, "")
}

// UnmarshalFieldWithTag unmarshals a field with optional struct tag information
func UnmarshalFieldWithTag(field reflect.Value, data ns.ByteArray, tag string) (int, error) {
	// handle pointer fields
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	length := parseLengthFromTag(tag)
	if length > 0 {
		if strings.Contains(field.Type().String(), "PrefixedOptional") &&
			strings.Contains(field.Type().String(), "FixedByteArray") {
			return unmarshalPrefixedOptionalFixedByteArray(field, data, length)
		}

		if strings.Contains(field.Type().String(), "Optional") &&
			strings.Contains(field.Type().String(), "FixedByteArray") {
			return unmarshalOptionalFixedByteArray(field, data, length)
		}
	}

	// has FromBytes method?
	if field.CanAddr() {
		if method := field.Addr().MethodByName("FromBytes"); method.IsValid() {
			results := method.Call([]reflect.Value{reflect.ValueOf(data)})
			if len(results) == 2 {
				if !results[1].IsNil() {
					err := results[1].Interface().(error)
					if strings.Contains(err.Error(), "does not implement FromBytes method") && field.Kind() == reflect.Slice {
						return unmarshalSliceWithReflection(field, data)
					}
					typeName := field.Type().String()
					if strings.Contains(typeName, "Optional[") && !strings.Contains(typeName, "PrefixedOptional") {
						if presentField := field.FieldByName("Present"); presentField.IsValid() && presentField.CanSet() {
							presentField.SetBool(false)
							return 0, nil
						}
					}
					return 0, err
				}
				return results[0].Interface().(int), nil
			}
		}
	}

	// if not addressable, try on the value itself
	if method := field.MethodByName("FromBytes"); method.IsValid() {
		results := method.Call([]reflect.Value{reflect.ValueOf(data)})
		if len(results) == 2 {
			if !results[1].IsNil() {
				err := results[1].Interface().(error)
				if strings.Contains(err.Error(), "does not implement FromBytes method") && field.Kind() == reflect.Slice {
					return unmarshalSliceWithReflection(field, data)
				}
				typeName := field.Type().String()
				if strings.Contains(typeName, "Optional[") && !strings.Contains(typeName, "PrefixedOptional") {
					return 0, nil
				}
				return 0, err
			}
			return results[0].Interface().(int), nil
		}
	}

	// handle other types
	switch field.Kind() {
	case reflect.Struct:
		return unmarshalStruct(field, data)

	case reflect.Slice:
		var length ns.VarInt
		n, err := length.FromBytes(data)
		if err != nil {
			return 0, err
		}
		offset := n

		slice := reflect.MakeSlice(field.Type(), int(length), int(length))
		for j := range int(length) {
			bytesConsumed, err := UnmarshalField(slice.Index(j), data[offset:])
			if err != nil {
				return offset, err
			}
			offset += bytesConsumed
		}
		field.Set(slice)
		return offset, nil

	case reflect.Array:
		offset := 0
		for j := range field.Len() {
			bytesConsumed, err := UnmarshalField(field.Index(j), data[offset:])
			if err != nil {
				return offset, err
			}
			offset += bytesConsumed
		}
		return offset, nil

	default:
		return 0, fmt.Errorf("unsupported type: %v", field.Type())
	}
}
