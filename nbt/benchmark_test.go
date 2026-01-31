package nbt_test

import (
	"testing"

	"github.com/go-mclib/protocol/nbt"
)

// Sample data structures for benchmarks

func makeSimpleCompound() nbt.Compound {
	return nbt.Compound{
		"name":  nbt.String("Steve"),
		"x":     nbt.Double(100.5),
		"y":     nbt.Double(64.0),
		"z":     nbt.Double(-200.5),
		"level": nbt.Int(42),
	}
}

func makeComplexCompound() nbt.Compound {
	items := make([]nbt.Tag, 36)
	for i := range items {
		items[i] = nbt.Compound{
			"id":    nbt.String("minecraft:diamond"),
			"count": nbt.Byte(64),
			"slot":  nbt.Byte(int8(i)),
		}
	}

	return nbt.Compound{
		"name":       nbt.String("Steve"),
		"x":          nbt.Double(100.5),
		"y":          nbt.Double(64.0),
		"z":          nbt.Double(-200.5),
		"yaw":        nbt.Float(90.0),
		"pitch":      nbt.Float(0.0),
		"onGround":   nbt.Byte(1),
		"health":     nbt.Float(20.0),
		"foodLevel":  nbt.Int(20),
		"xpLevel":    nbt.Int(30),
		"xpTotal":    nbt.Int(1395),
		"score":      nbt.Int(0),
		"dimension":  nbt.String("minecraft:overworld"),
		"playerUUID": nbt.IntArray{0x12345678, -0x65432110, 0x12345678, -0x65432110},
		"inventory": nbt.List{
			ElementType: nbt.TagCompound,
			Elements:    items,
		},
		"abilities": nbt.Compound{
			"flying":       nbt.Byte(0),
			"mayfly":       nbt.Byte(0),
			"instabuild":   nbt.Byte(0),
			"invulnerable": nbt.Byte(0),
			"walkSpeed":    nbt.Float(0.1),
			"flySpeed":     nbt.Float(0.05),
		},
	}
}

type SimplePlayer struct {
	Name  string  `nbt:"name"`
	X     float64 `nbt:"x"`
	Y     float64 `nbt:"y"`
	Z     float64 `nbt:"z"`
	Level int32   `nbt:"level"`
}

type Item struct {
	ID    string `nbt:"id"`
	Count int8   `nbt:"count"`
	Slot  int8   `nbt:"slot"`
}

type Abilities struct {
	Flying       bool    `nbt:"flying"`
	MayFly       bool    `nbt:"mayfly"`
	Instabuild   bool    `nbt:"instabuild"`
	Invulnerable bool    `nbt:"invulnerable"`
	WalkSpeed    float32 `nbt:"walkSpeed"`
	FlySpeed     float32 `nbt:"flySpeed"`
}

type ComplexPlayer struct {
	Name       string    `nbt:"name"`
	X          float64   `nbt:"x"`
	Y          float64   `nbt:"y"`
	Z          float64   `nbt:"z"`
	Yaw        float32   `nbt:"yaw"`
	Pitch      float32   `nbt:"pitch"`
	OnGround   bool      `nbt:"onGround"`
	Health     float32   `nbt:"health"`
	FoodLevel  int32     `nbt:"foodLevel"`
	XPLevel    int32     `nbt:"xpLevel"`
	XPTotal    int32     `nbt:"xpTotal"`
	Score      int32     `nbt:"score"`
	Dimension  string    `nbt:"dimension"`
	PlayerUUID []int32   `nbt:"playerUUID"`
	Inventory  []Item    `nbt:"inventory"`
	Abilities  Abilities `nbt:"abilities"`
}

// --- Encode Benchmarks ---

func BenchmarkEncodeSimple(b *testing.B) {
	compound := makeSimpleCompound()

	for b.Loop() {
		_, err := nbt.EncodeNetwork(compound)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeComplex(b *testing.B) {
	compound := makeComplexCompound()

	for b.Loop() {
		_, err := nbt.EncodeNetwork(compound)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeFile(b *testing.B) {
	compound := makeComplexCompound()

	for b.Loop() {
		_, err := nbt.EncodeFile(compound, "Player")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Decode Benchmarks ---

func BenchmarkDecodeSimple(b *testing.B) {
	compound := makeSimpleCompound()
	data, _ := nbt.EncodeNetwork(compound)

	for b.Loop() {
		_, err := nbt.DecodeNetwork(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeComplex(b *testing.B) {
	compound := makeComplexCompound()
	data, _ := nbt.EncodeNetwork(compound)

	for b.Loop() {
		_, err := nbt.DecodeNetwork(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeFile(b *testing.B) {
	compound := makeComplexCompound()
	data, _ := nbt.EncodeFile(compound, "Player")

	for b.Loop() {
		_, _, err := nbt.DecodeFile(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Marshal Benchmarks ---

func BenchmarkMarshalSimple(b *testing.B) {
	player := SimplePlayer{
		Name:  "Steve",
		X:     100.5,
		Y:     64.0,
		Z:     -200.5,
		Level: 42,
	}

	for b.Loop() {
		_, err := nbt.Marshal(player)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalComplex(b *testing.B) {
	items := make([]Item, 36)
	for i := range items {
		items[i] = Item{ID: "minecraft:diamond", Count: 64, Slot: int8(i)}
	}

	player := ComplexPlayer{
		Name:       "Steve",
		X:          100.5,
		Y:          64.0,
		Z:          -200.5,
		Yaw:        90.0,
		Pitch:      0.0,
		OnGround:   true,
		Health:     20.0,
		FoodLevel:  20,
		XPLevel:    30,
		XPTotal:    1395,
		Score:      0,
		Dimension:  "minecraft:overworld",
		PlayerUUID: []int32{0x12345678, -0x65432110, 0x12345678, -0x65432110},
		Inventory:  items,
		Abilities: Abilities{
			WalkSpeed: 0.1,
			FlySpeed:  0.05,
		},
	}

	for b.Loop() {
		_, err := nbt.Marshal(player)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Unmarshal Benchmarks ---

func BenchmarkUnmarshalSimple(b *testing.B) {
	player := SimplePlayer{
		Name:  "Steve",
		X:     100.5,
		Y:     64.0,
		Z:     -200.5,
		Level: 42,
	}
	data, _ := nbt.Marshal(player)

	for b.Loop() {
		var p SimplePlayer
		if err := nbt.Unmarshal(data, &p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalComplex(b *testing.B) {
	items := make([]Item, 36)
	for i := range items {
		items[i] = Item{ID: "minecraft:diamond", Count: 64, Slot: int8(i)}
	}

	player := ComplexPlayer{
		Name:       "Steve",
		X:          100.5,
		Y:          64.0,
		Z:          -200.5,
		Yaw:        90.0,
		Pitch:      0.0,
		OnGround:   true,
		Health:     20.0,
		FoodLevel:  20,
		XPLevel:    30,
		XPTotal:    1395,
		Score:      0,
		Dimension:  "minecraft:overworld",
		PlayerUUID: []int32{0x12345678, -0x65432110, 0x12345678, -0x65432110},
		Inventory:  items,
		Abilities: Abilities{
			WalkSpeed: 0.1,
			FlySpeed:  0.05,
		},
	}
	data, _ := nbt.Marshal(player)

	for b.Loop() {
		var p ComplexPlayer
		if err := nbt.Unmarshal(data, &p); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Round-trip Benchmarks ---

func BenchmarkRoundTripSimple(b *testing.B) {
	player := SimplePlayer{
		Name:  "Steve",
		X:     100.5,
		Y:     64.0,
		Z:     -200.5,
		Level: 42,
	}

	for b.Loop() {
		data, err := nbt.Marshal(player)
		if err != nil {
			b.Fatal(err)
		}
		var p SimplePlayer
		if err := nbt.Unmarshal(data, &p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTripComplex(b *testing.B) {
	items := make([]Item, 36)
	for i := range items {
		items[i] = Item{ID: "minecraft:diamond", Count: 64, Slot: int8(i)}
	}

	player := ComplexPlayer{
		Name:       "Steve",
		X:          100.5,
		Y:          64.0,
		Z:          -200.5,
		Yaw:        90.0,
		Pitch:      0.0,
		OnGround:   true,
		Health:     20.0,
		FoodLevel:  20,
		XPLevel:    30,
		XPTotal:    1395,
		Score:      0,
		Dimension:  "minecraft:overworld",
		PlayerUUID: []int32{0x12345678, -0x65432110, 0x12345678, -0x65432110},
		Inventory:  items,
		Abilities: Abilities{
			WalkSpeed: 0.1,
			FlySpeed:  0.05,
		},
	}

	for b.Loop() {
		data, err := nbt.Marshal(player)
		if err != nil {
			b.Fatal(err)
		}
		var p ComplexPlayer
		if err := nbt.Unmarshal(data, &p); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Allocation Benchmarks ---

func BenchmarkEncodeAllocations(b *testing.B) {
	compound := makeComplexCompound()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = nbt.EncodeNetwork(compound)
	}
}

func BenchmarkDecodeAllocations(b *testing.B) {
	compound := makeComplexCompound()
	data, _ := nbt.EncodeNetwork(compound)
	b.ReportAllocs()

	for b.Loop() {
		_, _ = nbt.DecodeNetwork(data)
	}
}

func BenchmarkMarshalAllocations(b *testing.B) {
	items := make([]Item, 36)
	for i := range items {
		items[i] = Item{ID: "minecraft:diamond", Count: 64, Slot: int8(i)}
	}

	player := ComplexPlayer{
		Name:      "Steve",
		X:         100.5,
		Y:         64.0,
		Z:         -200.5,
		Inventory: items,
	}
	b.ReportAllocs()

	for b.Loop() {
		_, _ = nbt.Marshal(player)
	}
}

func BenchmarkUnmarshalAllocations(b *testing.B) {
	items := make([]Item, 36)
	for i := range items {
		items[i] = Item{ID: "minecraft:diamond", Count: 64, Slot: int8(i)}
	}

	player := ComplexPlayer{
		Name:      "Steve",
		X:         100.5,
		Y:         64.0,
		Z:         -200.5,
		Inventory: items,
	}
	data, _ := nbt.Marshal(player)
	b.ReportAllocs()

	for b.Loop() {
		var p ComplexPlayer
		_ = nbt.Unmarshal(data, &p)
	}
}
