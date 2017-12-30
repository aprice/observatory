package utils

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"github.com/satori/go.uuid"
)

// NewTimeUUID generates a UUID-style 16-byte identifier, with the first 8 bytes
// being the current time in nanoseconds since the Unix epoch, and the last 8
// bytes being randomly set, ensuring that the RFC4122 variant bit is not set.
// These can be sorted and compared lexically as if they were a timestamp, but
// remain resilient against collisions and do not expose MAC addess or other
// host information.
func NewTimeUUID() uuid.UUID {
	id := TimeUUID(time.Now())
	rand.Read(id[8:])
	// Ensure we don't masquerade as MS GUID or RFC4122 UUID - clear variant bits
	id[8] = id[8] &^ 0xc0
	return id
}

// TimeUUID generates a UUID-style 16-byte identifier, with the first 8 bytes
// being the given Time's nanoseconds since the Unix epoch, and the last 8 bytes
// being zero. This can be used for lexical comparison of time-based UUIDs.
func TimeUUID(t time.Time) uuid.UUID {
	id := uuid.UUID{}
	binary.BigEndian.PutUint64(id[0:8], uint64(t.UnixNano()))
	return id
}
