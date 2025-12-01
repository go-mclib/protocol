package java_protocol_test

import (
	"bytes"
	"testing"

	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// Test struct for length tag functionality
type TestPacketWithLengthTags struct {
	// Test PrefixedOptional[FixedByteArray] with length tag
	Signature ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:256"`
	// Test Optional[FixedByteArray] with length tag
	OptionalSignature ns.Optional[ns.FixedByteArray] `mc:"length:32"`
	// Test different length
	ShortSignature ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:16"`
}

type TestPacket struct {
	Position ns.Position
	Active   ns.Boolean
	Score    ns.VarInt
}

type TestPacketWithTags struct {
	FixedInt   ns.Int
	VarInt     ns.VarInt
	CustomType ns.VarInt
	SkipField  ns.String `mc:"-"` // should be skipped
}

const (
	EnumA = iota
	EnumB
	EnumC
)

func TestAutomaticMarshalUnmarshal(t *testing.T) {
	original := TestPacket{
		Position: ns.Position{X: 100, Y: 64, Z: -200},
		Active:   ns.Boolean(true),
		Score:    ns.VarInt(12345),
	}

	// marshal
	data, err := jp.PacketDataToBytes(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// unmarshal
	var result TestPacket
	err = jp.BytesToPacketData(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// verify
	if result.Position != original.Position {
		t.Errorf("Position mismatch: got %v, want %v", result.Position, original.Position)
	}
	if result.Active != original.Active {
		t.Errorf("Active mismatch: got %v, want %v", result.Active, original.Active)
	}
	if result.Score != original.Score {
		t.Errorf("Score mismatch: got %v, want %v", result.Score, original.Score)
	}
}

func TestStructTags(t *testing.T) {
	original := TestPacketWithTags{
		FixedInt:   ns.Int(987654321),    // will be marshaled as 4-byte fixed int
		VarInt:     ns.VarInt(12345),     // will be marshaled as VarInt
		CustomType: ns.VarInt(EnumB),     // must convert enum to VarInt explicitly
		SkipField:  ns.String("ignored"), // should be ignored
	}

	// marshal
	data, err := jp.PacketDataToBytes(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// unmarshal
	var result TestPacketWithTags
	result.SkipField = ns.String("should_remain") // this should not be overwritten

	err = jp.BytesToPacketData(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// verify
	if result.FixedInt != original.FixedInt {
		t.Errorf("FixedInt mismatch: got %v, want %v", result.FixedInt, original.FixedInt)
	}
	if result.VarInt != original.VarInt {
		t.Errorf("VarInt mismatch: got %v, want %v", result.VarInt, original.VarInt)
	}
	if result.CustomType != original.CustomType {
		t.Errorf("CustomType mismatch: got %v, want %v", result.CustomType, original.CustomType)
	}
	// SkipField should remain unchanged
	if result.SkipField != ns.String("should_remain") {
		t.Errorf("SkipField was modified: got %v, want %v", result.SkipField, ns.String("should_remain"))
	}
}

func TestCompareWithManualMarshal(t *testing.T) {
	packet := TestPacket{
		Position: ns.Position{X: 1, Y: 2, Z: 3},
		Active:   ns.Boolean(true),
		Score:    ns.VarInt(999),
	}

	// marshal
	autoData, err := jp.PacketDataToBytes(packet)
	if err != nil {
		t.Fatalf("Automatic marshal failed: %v", err)
	}

	// marshal manually to compare
	manualData := ns.ByteArray{}

	posBytes, _ := packet.Position.ToBytes()
	manualData = append(manualData, posBytes...)

	activeBytes, _ := packet.Active.ToBytes()
	manualData = append(manualData, activeBytes...)

	scoreBytes, _ := packet.Score.ToBytes()
	manualData = append(manualData, scoreBytes...)

	// compare
	if !bytes.Equal(autoData, manualData) {
		t.Errorf("Automatic marshal differs from manual marshal")
		t.Logf("Auto:   %x", autoData)
		t.Logf("Manual: %x", manualData)
	}
}

func TestSliceHandling(t *testing.T) {
	type SlicePacket struct {
		Items []ns.String `mc:""`
		Nums  []ns.VarInt `mc:""`
	}

	original := SlicePacket{
		Items: []ns.String{ns.String("item1"), ns.String("item2"), ns.String("item3")},
		Nums:  []ns.VarInt{ns.VarInt(10), ns.VarInt(20), ns.VarInt(30)},
	}

	// marshal
	data, err := jp.PacketDataToBytes(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// unmarshal
	var result SlicePacket
	err = jp.BytesToPacketData(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// verify
	if len(result.Items) != len(original.Items) {
		t.Errorf("Items length mismatch: got %d, want %d", len(result.Items), len(original.Items))
	}
	for i, item := range result.Items {
		if item != original.Items[i] {
			t.Errorf("Items[%d] mismatch: got %v, want %v", i, item, original.Items[i])
		}
	}

	if len(result.Nums) != len(original.Nums) {
		t.Errorf("Nums length mismatch: got %d, want %d", len(result.Nums), len(original.Nums))
	}
	for i, num := range result.Nums {
		if num != original.Nums[i] {
			t.Errorf("Nums[%d] mismatch: got %v, want %v", i, num, original.Nums[i])
		}
	}
}

func TestNestedStructs(t *testing.T) {
	type InnerStruct struct {
		Value ns.String `mc:""`
		Count ns.VarInt `mc:""`
	}

	type OuterStruct struct {
		ID    ns.VarInt   `mc:""`
		Inner InnerStruct `mc:""`
		Name  ns.String   `mc:""`
	}

	original := OuterStruct{
		ID: ns.VarInt(123),
		Inner: InnerStruct{
			Value: ns.String("nested"),
			Count: ns.VarInt(456),
		},
		Name: ns.String("outer"),
	}

	// marshal
	data, err := jp.PacketDataToBytes(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// unmarshal
	var result OuterStruct
	err = jp.BytesToPacketData(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// verify
	if result.ID != original.ID {
		t.Errorf("ID mismatch: got %v, want %v", result.ID, original.ID)
	}
	if result.Inner.Value != original.Inner.Value {
		t.Errorf("Inner.Value mismatch: got %v, want %v", result.Inner.Value, original.Inner.Value)
	}
	if result.Inner.Count != original.Inner.Count {
		t.Errorf("Inner.Count mismatch: got %v, want %v", result.Inner.Count, original.Inner.Count)
	}
	if result.Name != original.Name {
		t.Errorf("Name mismatch: got %v, want %v", result.Name, original.Name)
	}
}

func BenchmarkAutomaticMarshal(b *testing.B) {
	packet := TestPacket{
		Position: ns.Position{X: 100, Y: 64, Z: -200},
		Active:   ns.Boolean(true),
		Score:    ns.VarInt(12345),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jp.PacketDataToBytes(packet)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAutomaticUnmarshal(b *testing.B) {
	packet := TestPacket{
		Position: ns.Position{X: 100, Y: 64, Z: -200},
		Active:   ns.Boolean(true),
		Score:    ns.VarInt(12345),
	}

	data, err := jp.PacketDataToBytes(packet)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result TestPacket
		err := jp.BytesToPacketData(data, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestPrefixedOptionalFixedByteArrayWithLength(t *testing.T) {
	t.Run("present with 256 bytes", func(t *testing.T) {
		type TestSingleSignature struct {
			Signature ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:256"`
		}

		testData := make(ns.ByteArray, 1+256)
		testData[0] = 1 // present = true
		for i := 1; i <= 256; i++ {
			testData[i] = byte(i % 256)
		}

		var packet TestSingleSignature
		err := jp.BytesToPacketData(testData, &packet)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !packet.Signature.Present {
			t.Error("Expected Signature.Present to be true")
		}

		if len(packet.Signature.Value.Data) != 256 {
			t.Errorf("Expected signature data length 256, got %d", len(packet.Signature.Value.Data))
		}

		for i, b := range packet.Signature.Value.Data {
			expected := byte((i + 1) % 256)
			if b != expected {
				t.Errorf("Signature data mismatch at index %d: got %02x, want %02x", i, b, expected)
				break
			}
		}
	})

	t.Run("not present", func(t *testing.T) {
		type TestSingleSignature struct {
			Signature ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:256"`
		}

		testData := ns.ByteArray{0}

		var packet TestSingleSignature
		err := jp.BytesToPacketData(testData, &packet)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if packet.Signature.Present {
			t.Error("Expected Signature.Present to be false")
		}
	})
}

func TestMultipleFieldsWithDifferentLengths(t *testing.T) {
	type TestMultipleLengths struct {
		Field1 ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:8"`
		Field2 ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:16"`
		Field3 ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:4"`
	}

	testData := make(ns.ByteArray, 3+8+16+4) // 3 presence bytes + data
	offset := 0

	testData[offset] = 1
	offset++
	for i := range 8 {
		testData[offset+i] = byte(0x10 + i)
	}
	offset += 8

	testData[offset] = 1
	offset++
	for i := range 16 {
		testData[offset+i] = byte(0x20 + i)
	}
	offset += 16

	testData[offset] = 1
	offset++
	for i := range 4 {
		testData[offset+i] = byte(0x30 + i)
	}

	var packet TestMultipleLengths
	err := jp.BytesToPacketData(testData, &packet)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !packet.Field1.Present {
		t.Error("Expected Field1.Present to be true")
	}
	if len(packet.Field1.Value.Data) != 8 {
		t.Errorf("Expected Field1 length 8, got %d", len(packet.Field1.Value.Data))
	}

	if !packet.Field2.Present {
		t.Error("Expected Field2.Present to be true")
	}
	if len(packet.Field2.Value.Data) != 16 {
		t.Errorf("Expected Field2 length 16, got %d", len(packet.Field2.Value.Data))
	}

	if !packet.Field3.Present {
		t.Error("Expected Field3.Present to be true")
	}
	if len(packet.Field3.Value.Data) != 4 {
		t.Errorf("Expected Field3 length 4, got %d", len(packet.Field3.Value.Data))
	}

	for i, b := range packet.Field1.Value.Data {
		expected := byte(0x10 + i)
		if b != expected {
			t.Errorf("Field1 data mismatch at index %d: got %02x, want %02x", i, b, expected)
		}
	}
}

// Test new conditional field feature
func TestConditionalFields(t *testing.T) {
	type TestConditionalPacket struct {
		MessageID ns.VarInt
		// Signature is only present when MessageID is 0
		Signature ns.Optional[ns.FixedByteArray] `mc:"if:MessageID,length:256"`
	}

	t.Run("signature present when MessageID is 0", func(t *testing.T) {
		original := TestConditionalPacket{
			MessageID: ns.VarInt(0),
			Signature: ns.Optional[ns.FixedByteArray]{
				Present: true,
				Value:   ns.FixedByteArray{Data: make(ns.ByteArray, 256), Length: 256},
			},
		}
		// Fill signature with test data
		for i := range original.Signature.Value.Data {
			original.Signature.Value.Data[i] = byte(i % 256)
		}

		// Marshal
		data, err := jp.PacketDataToBytes(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		// Unmarshal
		var result TestConditionalPacket
		err = jp.BytesToPacketData(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Verify
		if result.MessageID != 0 {
			t.Errorf("Expected MessageID 0, got %d", result.MessageID)
		}
		if !result.Signature.Present {
			t.Error("Expected Signature.Present to be true when MessageID is 0")
		}
		if len(result.Signature.Value.Data) != 256 {
			t.Errorf("Expected signature length 256, got %d", len(result.Signature.Value.Data))
		}
	})

	t.Run("signature not present when MessageID is non-zero", func(t *testing.T) {
		original := TestConditionalPacket{
			MessageID: ns.VarInt(5),
			Signature: ns.Optional[ns.FixedByteArray]{Present: false},
		}

		// Marshal
		data, err := jp.PacketDataToBytes(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		// Unmarshal
		var result TestConditionalPacket
		err = jp.BytesToPacketData(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Verify
		if result.MessageID != 5 {
			t.Errorf("Expected MessageID 5, got %d", result.MessageID)
		}
		if result.Signature.Present {
			t.Error("Expected Signature.Present to be false when MessageID is non-zero")
		}
	})
}

// Test FixedBitSet with length tag
func TestFixedBitSetWithLength(t *testing.T) {
	type TestBitSetPacket struct {
		Flags ns.FixedBitSet `mc:"length:20"`
	}

	original := TestBitSetPacket{
		Flags: ns.FixedBitSet{
			Length: 20,
			Data:   []byte{0xFF, 0x00, 0xAA}, // 3 bytes for 20 bits
		},
	}

	// Marshal
	data, err := jp.PacketDataToBytes(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var result TestBitSetPacket
	err = jp.BytesToPacketData(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify
	if result.Flags.Length != 20 {
		t.Errorf("Expected Length 20, got %d", result.Flags.Length)
	}
	if len(result.Flags.Data) != 3 {
		t.Errorf("Expected 3 bytes of data, got %d", len(result.Flags.Data))
	}
	if result.Flags.Data[0] != 0xFF || result.Flags.Data[1] != 0x00 || result.Flags.Data[2] != 0xAA {
		t.Errorf("Data mismatch: got %v, want [0xFF 0x00 0xAA]", result.Flags.Data)
	}
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	type SimplePacket struct {
		ID   ns.VarInt
		Name ns.String
	}

	t.Run("MarshalPacket", func(t *testing.T) {
		data := SimplePacket{
			ID:   ns.VarInt(42),
			Name: ns.String("test"),
		}

		packet, err := jp.MarshalPacket(jp.StatePlay, jp.C2S, 0x10, data)
		if err != nil {
			t.Fatalf("MarshalPacket failed: %v", err)
		}

		if packet.State != jp.StatePlay {
			t.Errorf("Expected StatePlay, got %v", packet.State)
		}
		if packet.Bound != jp.C2S {
			t.Errorf("Expected C2S, got %v", packet.Bound)
		}
		if packet.PacketID != 0x10 {
			t.Errorf("Expected PacketID 0x10, got %v", packet.PacketID)
		}
	})

	t.Run("UnmarshalPacket", func(t *testing.T) {
		original := SimplePacket{
			ID:   ns.VarInt(99),
			Name: ns.String("hello"),
		}

		packet, err := jp.MarshalPacket(jp.StateLogin, jp.S2C, 0x05, original)
		if err != nil {
			t.Fatalf("MarshalPacket failed: %v", err)
		}

		var result SimplePacket
		err = jp.UnmarshalPacket(packet, &result)
		if err != nil {
			t.Fatalf("UnmarshalPacket failed: %v", err)
		}

		if result.ID != original.ID {
			t.Errorf("ID mismatch: got %v, want %v", result.ID, original.ID)
		}
		if result.Name != original.Name {
			t.Errorf("Name mismatch: got %v, want %v", result.Name, original.Name)
		}
	})
}

// Test realistic packet: Chat message (C2S)
func TestRealisticChatPacket(t *testing.T) {
	type C2SChatData struct {
		Message      ns.String
		Timestamp    ns.Long
		Salt         ns.Long
		Signature    ns.PrefixedOptional[ns.FixedByteArray] `mc:"length:256"`
		MessageCount ns.VarInt
		Acknowledged ns.FixedBitSet `mc:"length:20"`
	}

	original := C2SChatData{
		Message:      ns.String("Hello, world!"),
		Timestamp:    ns.Long(1234567890),
		Salt:         ns.Long(9876543210),
		Signature:    ns.PrefixedOptional[ns.FixedByteArray]{Present: false},
		MessageCount: ns.VarInt(5),
		Acknowledged: ns.FixedBitSet{
			Length: 20,
			Data:   []byte{0x12, 0x34, 0x56},
		},
	}

	// Marshal
	data, err := jp.PacketDataToBytes(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var result C2SChatData
	err = jp.BytesToPacketData(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify
	if result.Message != original.Message {
		t.Errorf("Message mismatch: got %v, want %v", result.Message, original.Message)
	}
	if result.Timestamp != original.Timestamp {
		t.Errorf("Timestamp mismatch: got %v, want %v", result.Timestamp, original.Timestamp)
	}
	if result.Salt != original.Salt {
		t.Errorf("Salt mismatch: got %v, want %v", result.Salt, original.Salt)
	}
	if result.Signature.Present {
		t.Error("Expected Signature.Present to be false")
	}
	if result.MessageCount != original.MessageCount {
		t.Errorf("MessageCount mismatch: got %v, want %v", result.MessageCount, original.MessageCount)
	}
	if result.Acknowledged.Length != 20 {
		t.Errorf("Acknowledged.Length mismatch: got %v, want 20", result.Acknowledged.Length)
	}
}
