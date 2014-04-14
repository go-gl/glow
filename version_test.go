package main

import (
	"testing"
)

type versionTest struct {
	In  string
	Out Version
	Ok  bool
}

var versionTests = []versionTest{
	{"0.0", Version{0, 0}, true},
	{"0.1", Version{0, 1}, true},
	{"1.0", Version{1, 0}, true},
	{"5.6", Version{5, 6}, true},
	{"0.", Version{0, 0}, false},
	{"", Version{0, 0}, false},
	{".", Version{0, 0}, false},
	{"a", Version{0, 0}, false},
}

func TestVersion(t *testing.T) {
	for i := range versionTests {
		test := &versionTests[i]
		v, err := ParseVersion(test.In)
		if (err != nil) == test.Ok {
			t.Errorf("failed %v, %v", test, err)
		}
		if v.Valid() != test.Out.Valid() {
			t.Errorf("input != output %v", test)
		}
		if v.Compare(test.Out) != 0 {
			t.Errorf("input != output %v", test)
		}
	}
}
