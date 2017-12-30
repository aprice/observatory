package utils

import "testing"

func TestCompareSemVer(t *testing.T) {
	tests := map[string]struct {
		lhs         string
		rhs         string
		expected    int
		expectedErr error
	}{
		"blank-blank":     {"", "", 0, nil},
		"ok-blank":        {"1.0.0", "", 1, nil},
		"blank-ok":        {"", "1.0.0", -1, nil},
		"samerel":         {"1.0.0", "1.0.0", 0, nil},
		"patchup":         {"1.0.1", "1.0.0", 1, nil},
		"patchdown":       {"1.0.1", "1.0.2", -1, nil},
		"minorup":         {"1.1.0", "1.0.0", 1, nil},
		"minordown":       {"1.0.1", "1.1.0", -1, nil},
		"majorup":         {"2.0.0", "1.0.0", 1, nil},
		"majordown":       {"1.1.1", "2.0.0", -1, nil},
		"rel-dev":         {"1.0.1", "1.1.0-rc1", -1, nil},
		"dev-rel":         {"1.0.1-rc1", "1.0.1", -1, nil},
		"dev-dev":         {"1.0.1-rc1", "1.0.1-rc2", -1, nil},
		"ok-invalid":      {"1.0.0", "1.z.0", 0, ErrInvalidSemVer},
		"invalid-ok":      {"1.z.0", "1.0.0", 0, ErrInvalidSemVer},
		"invalid-invalid": {"1.z.0", "1.z.0", 0, nil},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			actual, err := CompareSemVer(test.lhs, test.rhs)
			if err != test.expectedErr || actual != test.expected {
				tt.Errorf("CompareSemVer(%q,%q) expected %d,%v, actual %d,%v",
					test.lhs, test.rhs,
					test.expected, test.expectedErr,
					actual, err)
			}
		})
	}
}
