package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// C2SIntentionPacket represents "Intention" (serverbound/handshake).
// > This packet causes the server to switch into the target state.
// It should be sent right after opening the TCP connection to prevent the server from disconnecting.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Handshake
var C2SIntentionPacket = jp.NewPacket(jp.StateHandshake, jp.C2S, 0x00)

const (
	IntentStatus ns.VarInt = iota + 1
	IntentLogin
	IntentTransfer
)

type C2SIntentionPacketData struct {
	ProtocolVersion ns.VarInt
	ServerAddress   ns.String
	ServerPort      ns.UnsignedShort
	Intent          ns.VarInt
}

// don't handle Legacy Server List Ping, as it's not part of
// the modern protocol that this library is designed to handle
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Legacy_Server_List_Ping
