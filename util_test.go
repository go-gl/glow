package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

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

type ByteSliceTest struct {
	in       []byte
	expected []byte
}

// The string needs a new line at the end.
var BlankLineTests = []ByteSliceTest{
	{nil, []byte("")},
	{[]byte(""), []byte("")},
	{[]byte("ä"), []byte("")},
	{[]byte("\n"), []byte("")},
	{[]byte("a\n"), []byte("a\n")},
	{[]byte("ä\n\n\n"), []byte("ä\n")},
	{[]byte("\nä\n\n"), []byte("ä\n")},
	{[]byte("\n\nä\n"), []byte("ä\n")},
	{[]byte("\n\n\nä"), []byte("")},
	{[]byte("\n//glow:keepspace\n\nä\n\n//glow:rmspace\n\n"), []byte("\nä\n\n")},
}

func TestBlankLineStrippingWriter(t *testing.T) {
	for n, test := range BlankLineTests {
		var out bytes.Buffer
		//w := &out
		w := NewBlankLineStrippingWriter(&out)
		l, err := w.Write(test.in)
		if err != nil {
			t.Errorf("BlankLineStrippingWriter[%d](%v): %s", n, test.in, err)
		}
		if l != len(test.in) {
			t.Errorf("BlankLineStrippingWriter[%d]:lenght: got %d, want %d", n, l, len(test.in))
		}
		b, err := ioutil.ReadAll(&out)
		if err != nil {
			t.Errorf("BlankLineStrippingWriter[%d](%v): %s", n, test.in, err)
		}
		if !bytes.Equal(b, test.expected) {
			t.Errorf("BlankLineStrippingWriter[%d](%v): got '%v', want '%v'",
				n, test.in, b, test.expected)
		}
	}
}

func TestBlankLineStrippingWriter2(t *testing.T) {
	const repeat = 20
	var out bytes.Buffer
	w := NewBlankLineStrippingWriter(&out)
	in := []byte(`Lorem ipsum dolor sit amet,
		consectetur adipiscing elit.


		Mauris aliquam metus id sagittis scelerisque. `)

	want := bytes.Repeat([]byte(`Lorem ipsum dolor sit amet,
		consectetur adipiscing elit.
		Mauris aliquam metus id sagittis scelerisque. `), repeat)
	want = append(want, '\n')

	for i := 0; i < repeat; i++ {
		if _, err := w.Write(in); err != nil {
			t.Errorf("BlankLineStrippingWriter2: %s", err)
		}
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		t.Errorf("BlankLineStrippingWriter2: %s", err)
	}
	b, err := ioutil.ReadAll(&out)
	if err != nil {
		t.Errorf("BlankLineStrippingWriter2: %s", err)
	}
	if !bytes.Equal(b, want) {
		t.Errorf("BlankLineStrippingWriter2: got\n'%s'...\nwant\n'%s'...\n",
			string(b[:50]), string(want[:50]))
	}
}
