package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// C2SKeepAlivePlayPacket represents "Clientbound Keep Alive (play)"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Clientbound_Keep_Alive_(play)
var C2SKeepAlivePlayPacket = jp.NewPacket(jp.StatePlay, jp.C2S, 0x1B)

type C2SKeepAlivePlayPacketData struct {
	KeepAliveID ns.Long
}

// C2SPingResponsePlayPacket represents "Ping Response (play)"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Ping_Response_(play)
var C2SPingResponsePlayPacket = jp.NewPacket(jp.StatePlay, jp.C2S, 0x18)

type C2SPingResponsePlayPacketData struct {
	ID ns.Int
}

// C2SChatMessagePacket represents "Chat Message" (unsigned)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Chat_Message
//
// Note: for low-level library, we expose raw content only; signing chain is handled at a higher layer as per project goals.
var C2SChatMessagePacket = jp.NewPacket(jp.StatePlay, jp.C2S, 0x03)

type C2SChatMessagePacketData struct {
	Message ns.String
}

// C2STeleportConfirmPacket represents "Teleport Confirm" (serverbound/play)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Teleport_Confirm
var C2STeleportConfirmPacket = jp.NewPacket(jp.StatePlay, jp.C2S, 0x00)

type C2STeleportConfirmPacketData struct {
	TeleportID ns.VarInt
}
