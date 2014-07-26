package main

import (
	"testing"
)

type parseTest struct {
	In  string
	Out Version
	Ok  bool
}

var parseTests = []parseTest{
	{"0.0", Version{0, 0}, true},
	{"0.1", Version{0, 1}, true},
	{"1.0", Version{1, 0}, true},
	{"5.6", Version{5, 6}, true},
	{"all", Version{-1, -1}, true},
	{"0.", Version{0, 0}, false},
	{"", Version{0, 0}, false},
	{".", Version{0, 0}, false},
	{"a", Version{0, 0}, false},
}

func TestParse(t *testing.T) {
	for i := range parseTests {
		test := &parseTests[i]
		v, err := ParseVersion(test.In)
		if (err != nil) == test.Ok {
			t.Errorf("failed %v, %v", test, err)
		}
		if v.Compare(test.Out) != 0 {
			t.Errorf("input != output %v", test)
		}
	}
}

func TestCompare(t *testing.T) {
	v10 := Version{1, 0}
	v11 := Version{1, 1}
	vAll := Version{-1, -1}
	if v10.Compare(v11) > 0 {
		t.Errorf("1.0 >= 1.1")
	}
	if v11.Compare(v10) < 0 {
		t.Errorf("1.1 >= 1.0")
	}
	if v10.Compare(v10) != 0 {
		t.Errorf("1.0 != 1.0")
	}
	if vAll.Compare(v10) != 0 || v10.Compare(vAll) != 0 {
		t.Errorf("1.0 != all")
	}
}
