package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// S2CStatusResponsePacket represents "Status Response" (clientbound/status).
// The response is a JSON string
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Status_Response
var S2CStatusResponsePacket = jp.NewPacket(jp.StateStatus, jp.S2C, 0x00)

type S2CStatusResponsePacketData struct {
	JSON ns.String
}

// S2CPongResponseStatusPacket represents "Pong Response (status)" (clientbound/status)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Pong_Response_(status)
var S2CPongResponseStatusPacket = jp.NewPacket(jp.StateStatus, jp.S2C, 0x01)

type S2CPongResponseStatusPacketData struct {
	Payload ns.Long
}
