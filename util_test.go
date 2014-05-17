package main

import "testing"

type trimAPIPrefixTest struct {
	in       string
	expected string
}

var trimAPIPrefixTests = []trimAPIPrefixTest{
	{"glTest", "Test"},
	{"wglTest", "Test"},
	{"eglTest", "Test"},
	{"glXTest", "Test"},
	{"GL_TEST", "TEST"},
	{"WGL_TEST", "TEST"},
	{"EGL_TEST", "TEST"},
	{"GLX_TEST", "TEST"},
	{"GL_0TEST", "GL_0TEST"},
	{"gl0Test", "gl0Test"},
}

func TestTrimApiPrefix(t *testing.T) {
	for _, test := range trimAPIPrefixTests {
		trimmed := TrimAPIPrefix(test.in)
		if trimmed != test.expected {
			t.Errorf("TrimAPIPrefix(%s) failed: %s != %s", test.in, test.expected, trimmed)
		}
	}
}
