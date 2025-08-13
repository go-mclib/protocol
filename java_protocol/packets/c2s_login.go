package packets

// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// C2SHelloPacket represents "Login Start" (serverbound/login).
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Hello
var C2SHelloPacket = jp.NewPacket(jp.StateLogin, jp.C2S, 0x00)

type C2SHelloPacketData struct {
	// Player's Username.
	Name ns.String
	// The UUID of the player logging in. Unused by the vanilla server.
	PlayerUUID ns.UUID
}

// C2SKeyPacket represents "Encryption Response" (serverbound/login).
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Encryption_Response
// https://minecraft.wiki/w/Protocol_encryption
var C2SKeyPacket = jp.NewPacket(jp.StateLogin, jp.C2S, 0x01)

type C2SKeyPacketData struct {
	// Shared Secret value, encrypted with the server's public key.
	SharedSecret ns.PrefixedByteArray
	// Verify Token value, encrypted with the same public key as the shared secret.
	VerifyToken ns.PrefixedByteArray
}

// C2SCustomQueryAnswerPacket represents "Login Plugin Response" (serverbound/login).
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Plugin_Response
var C2SCustomQueryAnswerPacket = jp.NewPacket(jp.StateLogin, jp.C2S, 0x02)

type C2SCustomQueryAnswerPacketData struct {
	// Should match ID from server.
	MessageID ns.VarInt
	// Any data, depending on the channel. The length of this array must be inferred
	// from the packet length. Only present if the client understood the request.
	Data ns.PrefixedOptional[ns.ByteArray]
}

// C2SLoginAcknowledgedPacket represents "Login Acknowledged" (serverbound/login). Has no fields
//
// > Acknowledgement to the Login Success packet sent by the server.
// This packet switches the connection state to configuration.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Plugin_Response
var C2SLoginAcknowledgedPacket = jp.NewPacket(jp.StateLogin, jp.C2S, 0x03)

// C2SCookieResponseLoginPacket represents "Cookie Response (login)" (serverbound/login).
//
// > Response to a Cookie Request (login) from the server.
// The vanilla server only accepts responses of up to 5 kiB in size.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Cookie_Response_(login)
var C2SCookieResponseLoginPacket = jp.NewPacket(jp.StateLogin, jp.C2S, 0x04)

type C2SCookieResponsePacketData struct {
	// The identifier of the cookie.
	Key ns.Identifier
	// The data of the cookie.
	Payload ns.PrefixedOptional[ns.ByteArray]
}
