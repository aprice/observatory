package collections

import "github.com/aprice/observatory/utils"

// StringSet is a Set of unique strings.
type StringSet map[string]utils.Sentinel

// NewStringSet from a list of strings.
func NewStringSet(strs ...string) StringSet {
	ss := make(StringSet, len(strs))
	ss.Add(strs...)
	return ss
}

// Copy this StringSet to a new StringSet.
func (ss StringSet) Copy() StringSet {
	copy := make(StringSet, len(ss))
	for k, v := range ss {
		copy[k] = v
	}
	return copy
}

// ToArray converts this StringSet to a []string.
func (ss StringSet) ToArray() []string {
	arr := make([]string, len(ss))
	i := 0
	for k := range ss {
		arr[i] = k
		i++
	}
	return arr
}

// Add a string to this set.
func (ss StringSet) Add(strs ...string) {
	for _, str := range strs {
		ss[str] = utils.Nothing
	}
}

// Remove a string from this set.
func (ss StringSet) Remove(str string) {
	delete(ss, str)
}

// Contains returns true if the given string is in the set.
func (ss StringSet) Contains(str string) bool {
	_, exists := ss[str]
	return exists
}

// ContainsAny returns true if any of the given strings are in the set.
func (ss StringSet) ContainsAny(strs ...string) bool {
	for _, str := range strs {
		if ss.Contains(str) {
			return true
		}
	}
	return false
}
