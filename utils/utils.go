package utils

import (
	"fmt"
	"math"
	"time"
)

// Sentinel is an empty struct used for flags and signals.
type Sentinel struct{}

// SentinelChannel is a low-overhead signalling channel.
type SentinelChannel chan Sentinel

// Signal sends an object over the channel.
func (sc SentinelChannel) Signal() {
	sc <- Nothing
}

// Nothing represents an empty value, used for values in set-type maps.
var Nothing = Sentinel{}

// DistinctStrings returns the array of unique values among the given string arrays.
func DistinctStrings(arrays ...[]string) []string {
	existing := make(map[string]Sentinel)
	ret := []string{}

	for _, array := range arrays {
		for _, item := range array {
			if _, ok := existing[item]; !ok {
				ret = append(ret, item)
				existing[item] = Nothing
			}
		}
	}

	return ret
}

// StringToArgs converts a shell-style string into an argv-style array.
func StringToArgs(in string) []string {
	out := []string{}
	escape := false
	inSingleQuote := false
	inDoubleQuote := false
	currentParam := []rune{}

	for _, char := range in {
		appendChar := true
		breakParam := false
		paramEmpty := len(currentParam) == 0

		if escape {
			escape = false
		} else if char == '\\' {
			escape = true
			appendChar = false
		} else if char == '"' && !inSingleQuote {
			appendChar = false
			breakParam = !paramEmpty || inDoubleQuote
			inDoubleQuote = !inDoubleQuote
		} else if char == '\'' && !inDoubleQuote {
			appendChar = false
			breakParam = !paramEmpty || inSingleQuote
			inSingleQuote = !inSingleQuote
		} else if char == ' ' && !inSingleQuote && !inDoubleQuote {
			appendChar = false
			breakParam = !paramEmpty
		}

		if appendChar {
			currentParam = append(currentParam, char)
		}

		if breakParam {
			out = append(out, string(currentParam))
			currentParam = currentParam[:0]
		}
	}
	if len(currentParam) > 0 {
		out = append(out, string(currentParam))
	}
	return out
}

// LaterDate returns the later of two dates.
func LaterDate(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

// LatestDate returns the latest of the given dates.
func LatestDate(dates ...time.Time) time.Time {
	ret := time.Time{}
	for _, date := range dates {
		if date.After(ret) {
			ret = date
		}
	}
	return ret
}

/*
public static String humanReadableByteCount(long bytes, boolean si) {
    int unit = si ? 1000 : 1024;
    if (bytes < unit) return bytes + " B";
    int exp = (int) (Math.log(bytes) / Math.log(unit));
    String pre = (si ? "kMGTPE" : "KMGTPE").charAt(exp-1) + (si ? "" : "i");
    return String.format("%.1f %sB", bytes / Math.pow(unit, exp), pre);
}
*/

var siUnits = [...]string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

// HumanReadableBytesSI returns SI-style human-readable byte count representation.
func HumanReadableBytesSI(bytes int64, sigdig int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	}
	exp := math.Log2(float64(bytes))
	unit := int(exp / 10.0)
	ofUnit := float64(bytes) / math.Pow(2, float64(unit)*10.0)
	decDig := int(math.Max(0, float64(sigdig)-math.Floor(math.Log10(ofUnit)+1)))
	return fmt.Sprintf("%.*f%s", decDig, ofUnit, siUnits[unit])
}

var decimalUnits = [...]string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

// HumanReadableBytesDecimal returns decimal-style human-readable byte count
// representation.
func HumanReadableBytesDecimal(bytes int64, sigdig int) string {
	if bytes < 1000 {
		return fmt.Sprintf("%dB", bytes)
	}
	fb := float64(bytes)
	exp := math.Log10(fb)
	unit := math.Floor(exp / 3.0)
	ofUnit := fb / math.Pow(10, unit*3.0)
	decDig := int(math.Max(0, float64(sigdig)-math.Floor(math.Log10(ofUnit)+1)))
	return fmt.Sprintf("%.*f%s", decDig, ofUnit, decimalUnits[int(unit)])
}
