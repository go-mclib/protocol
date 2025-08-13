package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// S2CKeepAlivePlayPacket represents "Serverbound Keep Alive (play)"
//
// > The server will frequently send out a keep-alive, each containing a random ID.
// The client must respond with the same payload.
// If the client does not respond to a Keep Alive packet within 15 seconds after it was sent,
// the server kicks the client. Vice versa, if the server does not send any keep-alives for 20 seconds,
// the client will disconnect and yields a "Timed out" exception.
//
// > The vanilla server uses a system-dependent time in milliseconds to generate the keep alive ID value.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Serverbound_Keep_Alive_(play)
var S2CKeepAlivePlayPacket = jp.NewPacket(jp.StatePlay, jp.S2C, 0x26)

type S2CKeepAlivePlayPacketData struct {
	KeepAliveID ns.Long
}

// S2CSystemChatMessagePacket represents "System Chat Message"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#System_Chat_Message
var S2CSystemChatMessagePacket = jp.NewPacket(jp.StatePlay, jp.S2C, 0x62)

type S2CSystemChatMessagePacketData struct {
	Content ns.JSONTextComponent
	Overlay ns.Boolean
}

// S2CPingPlayPacket represents "Ping (play)"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Ping_(play)
var S2CPingPlayPacket = jp.NewPacket(jp.StatePlay, jp.S2C, 0x33)

type S2CPingPlayPacketData struct {
	ID ns.Int
}
