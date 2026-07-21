package crypto

import (
	"crypto/md5"

	ns "github.com/go-mclib/protocol/java_protocol/net_structures"
)

// OfflinePlayerUUID returns Java's name-based UUID for OfflinePlayer:<name>.
func OfflinePlayerUUID(name string) ns.UUID {
	hash := md5.Sum([]byte("OfflinePlayer:" + name))
	uuid := ns.UUID(hash)
	uuid[6] = uuid[6]&0x0f | 0x30
	uuid[8] = uuid[8]&0x3f | 0x80
	return uuid
}
