package utils

import (
	"reflect"
	"testing"
	"time"

	"github.com/satori/go.uuid"
)

func TestNewTimeUUID(t *testing.T) {
	result := NewTimeUUID()
	if result == uuid.Nil {
		t.Errorf("NewTimeUUID was nil")
	} else {
		t.Logf("NewTimeUUID: %v", result)
	}
}

//func TimeUUID(t time.Time) uuid.UUID
func TestTimeUUID(t *testing.T) {
	var tests = []struct {
		time     time.Time
		expected uuid.UUID
	}{
		{time.Unix(1467943447, 0), uuid.FromBytesOrNil([]byte{0x14, 0x5F, 0x2E, 0xC9, 0x9C, 0xCC, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})},
	}

	for _, tt := range tests {
		actual := TimeUUID(tt.time)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf("TimeUUID(%v): expected %v, actual %v",
				tt.time, tt.expected, actual)
		}
	}
}
