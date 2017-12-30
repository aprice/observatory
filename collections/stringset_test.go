package collections

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// func (ss StringSet) Copy() StringSet
func TestCopy(t *testing.T) {
	orig := StringSet{}
	orig.Add("foo", "bar", "baz")

	copy := orig.Copy()

	if !reflect.DeepEqual(copy, orig) {
		t.Errorf("Original and copy do not match.")
	}
}

// func (ss StringSet) ToArray() []string
func TestToArray(t *testing.T) {
	orig := StringSet{}
	orig.Add("foo", "bar", "baz")

	expected := []string{"foo", "bar", "baz"}
	actual := orig.ToArray()

	sort.Strings(expected)
	sort.Strings(actual)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Original (%s) and copy (%s) do not match.", strings.Join(expected, ","), strings.Join(actual, ","))
	}
}

func TestManipulation(t *testing.T) {
	set := StringSet{}
	set.Add("foo")
	if len(set) != 1 {
		t.Errorf("After adding one item, length was %d", len(set))
	}

	set.Add("bar", "baz")
	if len(set) != 3 {
		t.Errorf("After adding two items, length was %d", len(set))
	}

	set.Add("bar")
	if len(set) != 3 {
		t.Errorf("After adding duplicate item, length was %d", len(set))
	}

	set.Add("bar", "qux")
	if len(set) != 4 {
		t.Errorf("After adding partial duplicate, length was %d", len(set))
	}

	set.Remove("qux")
	if len(set) != 3 {
		t.Errorf("After removing item, length was %d", len(set))
	}

	set.Remove("xyzzy")
	if len(set) != 3 {
		t.Errorf("After removing nonexistent item, length was %d", len(set))
	}
}

func TestContains(t *testing.T) {
	set := StringSet{}
	set.Add("foo", "bar", "baz")

	if !set.Contains("foo") {
		t.Errorf("Couldn't find foo in set")
	}

	if set.Contains("xyzzy") {
		t.Errorf("Found xyzzy in set")
	}

	if !set.ContainsAny("foo", "xyzzy") {
		t.Errorf("Couldn't find partial match in set")
	}

	if set.ContainsAny("xyzzy", "qux") {
		t.Errorf("Found xyzzy or qux in set")
	}
}
