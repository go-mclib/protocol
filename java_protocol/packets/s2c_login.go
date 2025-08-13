package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// S2CDisconnectLoginPacket represents "Disconnect (login)"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Disconnect_(login)
var S2CDisconnectLoginPacket = jp.NewPacket(jp.StateLogin, jp.S2C, 0x00)

type S2CDisconnectLoginPacketData struct {
	Reason ns.JSONTextComponent
}

// S2CEncryptionRequestPacket represents "Encryption Request"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Encryption_Request
//
// Note: PublicKey and VerifyToken are VarInt-length-prefixed byte arrays in the protocol;
// here we model them as raw bytes; callers should encode/decode with appropriate length prefixes
// where necessary until helper types are introduced.
var S2CEncryptionRequestPacket = jp.NewPacket(jp.StateLogin, jp.S2C, 0x01)

type S2CEncryptionRequestPacketData struct {
	ServerID  ns.String
	PublicKey ns.ByteArray
	VerifyTok ns.ByteArray
}

// S2CLoginSuccessPacket represents "Login Success"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Success
var S2CLoginSuccessPacket = jp.NewPacket(jp.StateLogin, jp.S2C, 0x02)

type S2CLoginSuccessPacketData struct {
	UUID     ns.UUID
	Username ns.String
}

// S2CSetCompressionPacket represents "Set Compression"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Set_Compression
var S2CSetCompressionPacket = jp.NewPacket(jp.StateLogin, jp.S2C, 0x03)

type S2CSetCompressionPacketData struct {
	Threshold ns.VarInt
}

// S2CLoginPluginRequestPacket represents "Login Plugin Request"
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Plugin_Request
var S2CLoginPluginRequestPacket = jp.NewPacket(jp.StateLogin, jp.S2C, 0x04)

type S2CLoginPluginRequestPacketData struct {
	MessageID ns.VarInt
	Channel   ns.Identifier
	Data      ns.ByteArray
}
