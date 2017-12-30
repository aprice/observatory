package model

// CheckStateByPriority is a collection type used for sorting CheckStates by status & age.
type CheckStateByPriority []CheckState

func (s CheckStateByPriority) Len() int {
	return len(s)
}
func (s CheckStateByPriority) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s CheckStateByPriority) Less(i, j int) bool {
	if s[i].Status < s[j].Status {
		return true
	}
	if s[i].Status > s[j].Status {
		return false
	}
	return s[i].StatusChanged.Before(s[j].StatusChanged)
}
