package packets_test

import (
	"bytes"
	"testing"

	jp "github.com/go-mclib/protocol/java_protocol"
	ps "github.com/go-mclib/protocol/java_protocol/packets"
	ns "github.com/go-mclib/protocol/net_structures"
)

var testPackets = map[*jp.Packet][]byte{}

func TestPackets(t *testing.T) {
	// build one concrete test: C2S Client Information (configuration)
	pkt, err := ps.C2SClientInformationPacket.WithData(ps.C2SClientInformationPacketData{
		Locale:              ns.String("en_us"),
		ViewDistance:        ns.Byte(10),
		ChatMode:            ns.VarInt(0),
		ChatColors:          ns.Boolean(true),
		DisplayedSkinParts:  ns.UnsignedByte(0x7f),
		MainHand:            ns.VarInt(1),
		EnableTextFiltering: ns.Boolean(true),
		AllowServerListings: ns.Boolean(true),
	})
	if err != nil {
		t.Fatalf("failed to build packet: %v", err)
	}
	// Expect trailing Extra VarInt(0) byte and adjusted length (0x0F)
	expected := []byte{0x0F, 0x00, 0x05, 0x65, 0x6e, 0x5f, 0x75, 0x73, 0x0a, 0x00, 0x01, 0x7f, 0x01, 0x01, 0x01, 0x00}

	actual, err := pkt.ToBytes(-1)
	if err != nil {
		t.Errorf("Error marshalling packet: %v", err)
	}
	if !bytes.Equal(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
