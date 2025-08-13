package java_protocol

import (
	"fmt"
	"reflect"

	ns "github.com/go-mclib/protocol/net_structures"
)

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

	for i := 0; i < val.NumField(); i++ {
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
		for j := 0; j < length; j++ {
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
		for j := 0; j < field.Len(); j++ {
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

	for i := 0; i < val.NumField(); i++ {
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

		// is there data left?
		if offset >= len(data) {
			return offset, fmt.Errorf("insufficient data for field %s", fieldType.Name)
		}

		// unmarshal
		bytesConsumed, err := unmarshalField(field, data[offset:])
		if err != nil {
			return offset, fmt.Errorf("error unmarshaling field %s: %w", fieldType.Name, err)
		}

		offset += bytesConsumed
	}

	return offset, nil
}

func unmarshalField(field reflect.Value, data ns.ByteArray) (int, error) {
	// handle pointer fields
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	// has FromBytes method?
	if field.CanAddr() {
		if method := field.Addr().MethodByName("FromBytes"); method.IsValid() {
			results := method.Call([]reflect.Value{reflect.ValueOf(data)})
			if len(results) == 2 {
				if !results[1].IsNil() {
					return 0, results[1].Interface().(error)
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
				return 0, results[1].Interface().(error)
			}
			return results[0].Interface().(int), nil
		}
	}

	// handle other types
	switch field.Kind() {
	case reflect.Struct:
		// recursively unmarshal nested structs
		return unmarshalStruct(field, data)

	case reflect.Slice:
		// decode length first
		var length ns.VarInt
		n, err := length.FromBytes(data)
		if err != nil {
			return 0, err
		}
		offset := n

		// create slice
		slice := reflect.MakeSlice(field.Type(), int(length), int(length))
		for j := 0; j < int(length); j++ {
			bytesConsumed, err := unmarshalField(slice.Index(j), data[offset:])
			if err != nil {
				return offset, err
			}
			offset += bytesConsumed
		}
		field.Set(slice)
		return offset, nil

	case reflect.Array:
		// fixed-size arrays
		offset := 0
		for j := 0; j < field.Len(); j++ {
			bytesConsumed, err := unmarshalField(field.Index(j), data[offset:])
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
