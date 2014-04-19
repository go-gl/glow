package main

import "testing"

type trimGLEnumPrefixTest struct {
	in       string
	expected string
}

var trimGLEnumPrefixTests = []trimGLEnumPrefixTest{
	{"", ""},
	{"_", "_"},
	{"GL_", ""},
	{"GL_bla", "bla"},
	{"GL123", "GL123"},
	{"GL_0123", "GL_0123"},
	{"GL", "GL"},
}

func TestEnumPrefix(t *testing.T) {
	for i := range trimGLEnumPrefixTests {
		test := &trimGLEnumPrefixTests[i]
		trimmed := TrimGLEnumPrefix(test.in)
		if trimmed != test.expected {
			t.Errorf("TrimGLEnumPrefix(%s) failed: %s != %s", test.in, test.expected, trimmed)
		}
	}
}
