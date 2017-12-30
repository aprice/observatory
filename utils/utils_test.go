package utils

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDistinctStrings(t *testing.T) {
	var tests = []struct {
		in       []string
		expected []string
	}{
		{[]string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}},
		{[]string{"foo", "foo", "foo"}, []string{"foo"}},
		{[]string{"foo", "foo", "baz"}, []string{"foo", "baz"}},
		{[]string{"foo", "bar", "foo"}, []string{"foo", "bar"}},
		{[]string{"foo"}, []string{"foo"}},
		{[]string{}, []string{}},
	}

	for _, tt := range tests {
		actual := DistinctStrings(tt.in)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf("DistinctStrings(%v): expected %v, actual %v",
				tt.in, tt.expected, actual)
		}
	}
}

func TestStringToArgs(t *testing.T) {
	var tests = []struct {
		in       string
		expected []string
	}{
		{"foo bar baz", []string{"foo", "bar", "baz"}},
		{"foo  bar baz", []string{"foo", "bar", "baz"}},
		{"foo \"bar baz\" qux", []string{"foo", "bar baz", "qux"}},
		{"foo \"bar'baz\" qux", []string{"foo", "bar'baz", "qux"}},
		{"foo \"\" bar", []string{"foo", "", "bar"}},
		{"foo \\\"bar baz", []string{"foo", "\"bar", "baz"}},
		{"foo \\'bar baz", []string{"foo", "'bar", "baz"}},
		{"foo \\\\bar baz", []string{"foo", "\\bar", "baz"}},
		{"foo 'bar baz' qux", []string{"foo", "bar baz", "qux"}},
		{"foo 'bar\"baz' qux", []string{"foo", "bar\"baz", "qux"}},
		{"foo '' bar", []string{"foo", "", "bar"}},
		{"foo", []string{"foo"}},
	}

	for _, tt := range tests {
		actual := StringToArgs(tt.in)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf("StringToArgs(%v): expected %v, actual %v",
				tt.in, strings.Join(tt.expected, ","), strings.Join(actual, ","))
		}
	}
}

func TestLaterDate(t *testing.T) {
	var tests = []struct {
		t1       time.Time
		t2       time.Time
		expected time.Time
	}{
		{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		actual := LaterDate(tt.t1, tt.t2)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf("LaterDate(%v, %v): expected %v, actual %v",
				tt.t1, tt.t2, tt.expected, actual)
		}
	}
}

func TestLatestDate(t *testing.T) {
	var tests = []struct {
		in       []time.Time
		expected time.Time
	}{
		{[]time.Time{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local)}, time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{[]time.Time{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local), time.Date(1988, 3, 31, 0, 30, 0, 0, time.Local)}, time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{[]time.Time{time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)}, time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{[]time.Time{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)}, time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{[]time.Time{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)}, time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{[]time.Time{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local)}, time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
		{[]time.Time{time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local), time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local), time.Date(1983, 4, 29, 0, 30, 0, 0, time.Local)}, time.Date(2016, 7, 9, 0, 30, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		actual := LatestDate(tt.in...)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf("LatestDate(%v): expected %v, actual %v",
				tt.in, tt.expected, actual)
		}
	}
}

func TestHumanReadableBytesSI(t *testing.T) {
	var tests = []struct {
		bytes    int64
		sigdig   int
		expected string
	}{
		{5, 3, "5B"},
		{1000, 3, "1000B"},
		{1023, 3, "1023B"},
		{1024, 3, "1.00KiB"},
		{1024, 2, "1.0KiB"},
		{1024, 1, "1KiB"},
		{1124, 2, "1.1KiB"},
		{1124, 1, "1KiB"},
		{102400, 3, "100KiB"},
		{102400, 4, "100.0KiB"},
		{10240, 3, "10.0KiB"},
		{10240, 2, "10KiB"},
		{102400, 1, "100KiB"},
		{1048576, 1, "1MiB"},
		{1073741824, 1, "1GiB"},
		{1099511627776, 1, "1TiB"},
		{1125899906842624, 1, "1PiB"},
		{1152921504606846976, 1, "1EiB"},
	}

	for _, tt := range tests {
		actual := HumanReadableBytesSI(tt.bytes, tt.sigdig)
		if actual != tt.expected {
			t.Errorf("HumanReadableBytesSI(%v, %v): expected %v, actual %v",
				tt.bytes, tt.sigdig, tt.expected, actual)
		}
	}
}

func TestHumanReadableBytesDecimal(t *testing.T) {
	var tests = []struct {
		bytes    int64
		sigdig   int
		expected string
	}{
		{5, 3, "5B"},
		{999, 3, "999B"},
		{1000, 3, "1.00KB"},
		{1023, 3, "1.02KB"},
		{1024, 3, "1.02KB"},
		{1024, 2, "1.0KB"},
		{1024, 1, "1KB"},
		{1124, 2, "1.1KB"},
		{1124, 1, "1KB"},
		{100000, 3, "100KB"},
		{100000, 4, "100.0KB"},
		{100000, 1, "100KB"},
		{1000000, 1, "1MB"},
		{1000000000, 1, "1GB"},
		{1000000000000, 1, "1TB"},
		//BUG: float64 precision screws us over here somehow
		{1000000000000003, 1, "1PB"},
		{1000000000000000000, 1, "1EB"},
	}

	for _, tt := range tests {
		actual := HumanReadableBytesDecimal(tt.bytes, tt.sigdig)
		if actual != tt.expected {
			t.Errorf("HumanReadableBytesDecimal(%v, %v): expected %v, actual %v",
				tt.bytes, tt.sigdig, tt.expected, actual)
		}
	}
}
