package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// C2SStatusRequestPacket represents "Status Request" (serverbound/status). Has no fields.
//
// > The status can only be requested once immediately after the handshake, before any ping.
// The server won't respond otherwise.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Status_Request
var C2SStatusRequestPacket = jp.NewPacket(jp.StateStatus, jp.C2S, 0x00)

// C2SPingRequestPacket represents "Ping Request (status)" (serverbound/status)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Ping_Request_(status)
var C2SPingRequestPacket = jp.NewPacket(jp.StateStatus, jp.C2S, 0x01)

type C2SPingRequestPacketData struct {
	// May be any number, but vanilla clients will always use the timestamp in milliseconds.
	Timestamp ns.Long
}
