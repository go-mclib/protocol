package packets

import (
	jp "github.com/go-mclib/protocol/java_protocol"
	ns "github.com/go-mclib/protocol/net_structures"
)

// C2SClientInformationPacket represents "Client Information" (serverbound/configuration).
//
// > Sent when the player connects, or when settings are changed.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Client_Information_(configuration)
var C2SClientInformationPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x00)

type C2SClientInformationPacketData struct {
	// e. g. `en_GB`
	Locale ns.String
	// Client-side render distance, in chunks.
	ViewDistance ns.Byte
	// 0: enabled, 1: commands only, 2: hidden, see [ChatMode]
	ChatMode ns.VarInt
	// "Colors" multiplayer setting. The vanilla server stores this value but does nothing with it
	// (see [MC-64867](https://bugs.mojang.com/browse/MC/issues/MC-64867)).
	// Third-party servers such as Hypixel disable all coloring in chat and system messages when it is false.
	ChatColors ns.Boolean
	// Bit mask, see [DisplayedSkinParts]
	DisplayedSkinParts ns.UnsignedByte
	// 0: Left, 1: Right, see [MainHand]
	MainHand ns.VarInt
	// Enables filtering of text on signs and written book titles. The vanilla client sets this
	// according to the `profanityFilterPreferences.profanityFilterOn` account attribute indicated
	// by the `/player/attributes` Mojang API endpoint. In offline mode it is always false.
	EnableTextFiltering ns.Boolean
	// Servers usually list online players, this option should let you not show up in that list.
	AllowServerListings ns.Boolean
	// 0: all, 1: decreased, 2: minimal, see [ParticleStatus]
	ParticleStatus ns.VarInt
}

type ChatMode ns.VarInt

const (
	ChatModeEnabled ChatMode = iota
	ChatModeCommandsOnly
	ChatModeHidden
)

type DisplayedSkinParts struct {
	// 0x01 - Cape enabled
	Cape byte
	// 0x02 - Jacket enabled
	Jacket byte
	// 0x04 - Left Sleeve enabled
	LeftSleeve byte
	// 0x08 - Right Sleeve enabled
	RightSleeve byte
	// 0x10 - Left Pants Leg enabled
	LeftPantsLeg byte
	// 0x20 - Right Pants Leg enabled
	RightPantsLeg byte
	// 0x40 - Hat enabled
	Hat byte
	// The most significant bit (bit 7, 0x80) appears to be unused.
}

func (d *DisplayedSkinParts) FromBytes(b []byte) {
	d.Cape = b[0] & 0x01
	d.Jacket = b[0] & 0x02
	d.LeftSleeve = b[0] & 0x04
	d.RightSleeve = b[0] & 0x08
	d.LeftPantsLeg = b[0] & 0x10
	d.RightPantsLeg = b[0] & 0x20
	d.Hat = b[0] & 0x40
}

func (d *DisplayedSkinParts) ToBytes() []byte {
	return []byte{
		(d.Cape << 0) | (d.Jacket << 1) | (d.LeftSleeve << 2) | (d.RightSleeve << 3) |
			(d.LeftPantsLeg << 4) | (d.RightPantsLeg << 5) | (d.Hat << 6),
	}
}

type MainHand ns.VarInt

const (
	MainHandLeft MainHand = iota
	MainHandRight
)

type ParticleStatus ns.VarInt

const (
	ParticleStatusAll ParticleStatus = iota
	ParticleStatusDecreased
	ParticleStatusMinimal
)

// C2SCookieResponseConfigurationPacket represents "Cookie Response (configuration)" (serverbound/configuration).
//
// > Response to a Cookie Request (configuration) from the server.
// The vanilla server only accepts responses of up to 5 kiB in size.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Cookie_Response_(configuration)
var C2SCookieResponseConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x01)

type C2SCookieResponseConfigurationPacketData struct {
	// The identifier of the cookie.
	Key ns.Identifier
	// The data of the cookie.
	Payload ns.PrefixedOptional[ns.ByteArray]
}

// C2SCustomPayloadPacket represents "Serverbound Plugin Message (configuration)" (serverbound/configuration).
//
// > Mods and plugins can use this to send their data. Minecraft itself uses some plugin channels.
// These internal channels are in the minecraft namespace.
//
// > Note that the length of Data is known only from the packet length, since the packet has no length field of any kind.
// In vanilla server, the maximum data length is 32767 bytes.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Serverbound_Plugin_Message_(configuration)
var C2SCustomPayloadPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x02)

type C2SCustomPayloadPacketData struct {
	// Name of the plugin channel used to send the data.
	Channel ns.Identifier
	// Any data, depending on the channel.
	// `minecraft:` channels are documented [here](https://minecraft.wiki/w/Java_Edition_protocol/Plugin_channels).
	// The length of this array must be inferred from the packet length.
	Data ns.ByteArray
}

// C2SFinishConfigurationPacket represents "Acknowledge Finish Configuration"
//
// > Sent by the client to notify the server that the configuration process has finished.
// It is sent in response to the server's Finish Configuration.
// This packet switches the connection state to play.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Acknowledge_Finish_Configuration
var C2SFinishConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x03)

// C2SKeepAliveConfigurationPacket represents "Serverbound Keep Alive (configuration)"
//
// > The server will frequently send out a keep-alive packet, each containing a random ID.
// The client must respond with the same packet.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Serverbound_Keep_Alive_(configuration)
var C2SKeepAliveConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x04)

type C2SKeepAliveConfigurationPacketData struct {
	// A random ID sent by the server.
	KeepAliveID ns.Long
}

// C2SPongConfigurationPacket represents "Pong (configuration)"
//
// > Response to the clientbound packet (Ping) with the same id.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Pong_(configuration)
var C2SPongConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x05)

type C2SPongConfigurationPacketData struct {
	// The ID of the packet that the client received from the server.
	ID ns.Int
}

// C2SResourcePackConfigurationPacket represents "Resource Pack Response (Configuration)".
//
// > Sent by the client to the server to indicate how it handled a resource pack request.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Resource_Pack_Response_(Configuration)
var C2SResourcePackConfigurationPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x06)

type C2SResourcePackConfigurationPacketData struct {
	// The unique identifier of the resource pack received in the "Add Resource Pack (configuration)" request.
	UUID ns.UUID
	// Result ID, see [ResourcePackStatus]
	Result ns.VarInt
}

type ResourcePackStatus ns.VarInt

const (
	// Successfully downloaded
	ResourcePackStatusSuccessfullyDownloaded ResourcePackStatus = iota
	// Declined
	ResourcePackStatusDeclined
	// Failed to download
	ResourcePackStatusFailedToDownload
	// Accepted
	ResourcePackStatusAccepted
	// Downloaded
	ResourcePackStatusDownloaded
	// Invalid URL
	ResourcePackStatusInvalidURL
	// Failed to reload
	ResourcePackStatusFailedToReload
	// Discarded
	ResourcePackStatusDiscarded
)

// C2SSelectKnownPacksPacket represents "Serverbound Known Packs" (serverbound/configuration).
//
// > Informs the server of which data packs are present on the client.
// The client sends this in response to Clientbound Known Packs.
//
// > If the client specifies a pack in this packet, the server should omit its contained data from the Registry Data packet.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Serverbound_Known_Packs
var C2SSelectKnownPacksPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x07)

type C2SSelectKnownPacksPacketData struct {
	KnownPacks []KnownPack
}

type KnownPack struct {
	Namespace ns.String
	ID        ns.String
	Version   ns.String
}

// C2SCustomClickActionPacket represents "Custom Click Action (configuration)" packet.
//
// > Sent when the client clicks a Text Component with the `minecraft:custom` click action.
// This is meant as an alternative to running a command, but will not have any effect on vanilla servers.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Custom_Click_Action_(configuration)
var C2SCustomClickActionPacket = jp.NewPacket(jp.StateConfiguration, jp.C2S, 0x08)

type C2SCustomClickActionPacketData struct {
	// The identifier for the click action.
	ID ns.Identifier
	// The data to send with the click action. May be a `TAG_END` (0).
	Payload ns.NBT
}
