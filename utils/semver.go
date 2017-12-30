package utils

import (
	"errors"
	"strconv"
	"strings"
)

// ErrInvalidSemVer indicates a semver value that is not correctly formatted.
var ErrInvalidSemVer = errors.New("invalid semver format")

// CompareSemVer compares two SemVer values, and returns -1 if lhs is less
// (older) than rhs, 1 if lhs is greater (newer) than rhs, and 0 if they are
// equivalent. The error will be ErrInvalidSemVer if either value is not valid
// semver format, or nil otherwise.
func CompareSemVer(lhs, rhs string) (int, error) {
	// Simple cases
	if lhs == rhs {
		return 0, nil
	}
	if lhs == "" {
		return -1, nil
	}
	if rhs == "" {
		return 1, nil
	}

	lParts := strings.Split(lhs, "-")
	rParts := strings.Split(rhs, "-")

	if lParts[0] != rParts[0] {
		lSubParts := strings.Split(lParts[0], ".")
		rSubParts := strings.Split(rParts[0], ".")
		for i, lv := range lSubParts {
			rv := rSubParts[i]
			if lv == rv {
				continue
			}
			ilv, err := strconv.Atoi(lv)
			if err != nil {
				return 0, ErrInvalidSemVer
			}
			irv, err := strconv.Atoi(rv)
			if err != nil {
				return 0, ErrInvalidSemVer
			}
			if ilv > irv {
				return 1, nil
			}
			if ilv < irv {
				return -1, nil
			}
		}
	}

	clParts := len(lParts)
	crParts := len(rParts)

	if clParts == 1 && crParts > 1 {
		return 1, nil
	}
	if clParts > 1 && crParts == 1 {
		return -1, nil
	}

	if clParts > crParts {
		return 1, nil
	}
	if clParts < crParts {
		return -1, nil
	}

	for i, lv := range lParts {
		rv := rParts[i]
		if lv == rv {
			continue
		}
		ilv, err := strconv.Atoi(lv)
		irv, err2 := strconv.Atoi(rv)
		if err != nil || err2 != nil {
			if lv > rv {
				return 1, nil
			}
			if lv < rv {
				return -1, nil
			}
		} else {
			if ilv > irv {
				return 1, nil
			}
			if ilv < irv {
				return -1, nil
			}
		}
	}
	return 0, nil
}
