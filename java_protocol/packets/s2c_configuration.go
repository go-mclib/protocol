package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// S2CFinishConfigurationPacket represents "Finish Configuration".
// Has no data
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Finish_Configuration
var S2CFinishConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.S2C, 0x03)

// S2CKeepAliveConfigurationPacket represents "Clientbound Keep Alive (configuration)"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Clientbound_Keep_Alive_(configuration)
var S2CKeepAliveConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.S2C, 0x04)

type S2CKeepAliveConfigurationPacketData struct {
	ID ns.Long
}

// S2CPingConfigurationPacket represents "Ping (configuration)"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Ping_(configuration)
var S2CPingConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.S2C, 0x05)

type S2CPingConfigurationPacketData struct {
	ID ns.Int
}
