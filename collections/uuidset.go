package collections

import (
	"github.com/aprice/observatory/utils"
	"github.com/satori/go.uuid"
)

// UUIDSet is a Set of unique UUIDs.
type UUIDSet map[uuid.UUID]utils.Sentinel

// Copy this UUIDSet to a new UUIDSet.
func (us UUIDSet) Copy() UUIDSet {
	copy := make(UUIDSet, len(us))
	for k, v := range us {
		copy[k] = v
	}
	return copy
}

// ToArray converts this UUIDSet to a []uuid.UUID.
func (us UUIDSet) ToArray() []uuid.UUID {
	arr := make([]uuid.UUID, len(us))
	i := 0
	for k := range us {
		arr[i] = k
		i++
	}
	return arr
}

// Add a UUID to this set.
func (us UUIDSet) Add(uuids ...uuid.UUID) {
	for _, id := range uuids {
		us[id] = utils.Nothing
	}
}

// Remove a UUID from this set.
func (us UUIDSet) Remove(id uuid.UUID) {
	delete(us, id)
}

// Contains returns true if the given UUID is in the set.
func (us UUIDSet) Contains(id uuid.UUID) bool {
	_, exists := us[id]
	return exists
}

// ContainsAny returns true if any of the given UUIDs are in the set.
func (us UUIDSet) ContainsAny(uuids ...uuid.UUID) bool {
	for _, id := range uuids {
		if us.Contains(id) {
			return true
		}
	}
	return false
}
