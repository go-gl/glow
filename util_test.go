package main

import "testing"

type trimApiPrefixTest struct {
	in       string
	expected string
}

var trimApiPrefixTests = []trimApiPrefixTest{
	{"glTest", "Test"},
	{"wglTest", "Test"},
	{"eglTest", "Test"},
	{"glxTest", "Test"},
	{"GL_TEST", "TEST"},
	{"WGL_TEST", "TEST"},
	{"EGL_TEST", "TEST"},
	{"GLX_TEST", "TEST"},
	{"GL_0TEST", "GL_0TEST"},
	{"gl0Test", "gl0Test"},
}

func TestTrimApiPrefix(t *testing.T) {
	for _, test := range trimApiPrefixTests {
		trimmed := TrimApiPrefix(test.in)
		if trimmed != test.expected {
			t.Errorf("TrimApiPrefix(%s) failed: %s != %s", test.in, test.expected, trimmed)
		}
	}
}
