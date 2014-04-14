package main

import (
	"testing"
)

type testEnumPrefix struct {
	In  string
	Out string
}

type testCamelCase struct {
	In  string
	Out string
}

var allTestsEnumPrefix = []testEnumPrefix{
	{"", ""},
	{"_", "_"},
	{"GL_", ""},
	{"GL_bla", "bla"},
	{"GL123", "GL123"},
	{"GL_0123", "GL_0123"},
	{"GL", "GL"},
}

var allTestsCamelCase = []testCamelCase{
	{"", ""},
	{"_", ""},
	{"ARB_multisample", "ARBMultisample"},
	{"vertex_array_object", "VertexArrayObject"},
	{"a_b_c_", "ABC"},
	{"ABC", "ABC"},
	{"1_2_", "12"},
}

func TestCamelCase(t *testing.T) {
	for i := range allTestsCamelCase {
		te := &allTestsCamelCase[i]
		cc := CamelCase(te.In)
		if cc != te.Out {
			t.Errorf("CamelCase() failed: %s -> %s (%s != %s)", te.In, te.Out, cc, te.Out)
		}
	}
}

func TestEnumPrefix(t *testing.T) {
	for i := range allTestsEnumPrefix {
		te := &allTestsEnumPrefix[i]
		tr := TrimGLEnumPrefix(te.In)
		if tr != te.Out {
			t.Errorf("TrimGLEnumPrefix() failed: %s -> %s (%s != %s)", te.In, te.Out, tr, te.Out)
		}
	}
}
