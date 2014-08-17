package main

import (
	"bytes"
	"io"
	"strings"
	"unicode"
)

// TrimAPIPrefix removes the API-specific prefix from a spec name.
// e.g., glTest becomes Test; GLX_TEST becomes TEST; egl0Test stays egl0Test
func TrimAPIPrefix(name string) string {
	prefixes := []string{"glX", "wgl", "egl", "gl", "GLX_", "WGL_", "EGL_", "GL_"}

	trimmed := name
	prefix := ""
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			trimmed = strings.TrimPrefix(name, p)
			prefix = p
			break
		}
	}

	if strings.IndexAny(trimmed, "0123456789") == 0 {
		return prefix + trimmed
	}
	return trimmed
}

// BlankLineStrippingWriter removes whitespace- or comment-only lines delimited
// by \n. A necessary evil to work around how text/template handles whitespace.
// The template needs a new line at the end.
type BlankLineStrippingWriter struct {
	output io.Writer
	buf    *bytes.Buffer
}

// NewBlankLineStrippingWriter creates a new BlankLineStrippingWriter.
func NewBlankLineStrippingWriter(wrapped io.Writer) *BlankLineStrippingWriter {
	return &BlankLineStrippingWriter{wrapped, new(bytes.Buffer)}
}

func isBlank(line string) bool {
	blank := true
	for _, ch := range line {
		if !unicode.IsSpace(ch) && ch != '/' {
			blank = false
			break
		}
	}
	return blank
}

// Write appends the contents of p to the BlankLineStrippingWriter.
// The return values are the length of p and the error of the underlaying io.Writer.
func (w BlankLineStrippingWriter) Write(p []byte) (int, error) {
	// Buffer the current write.
	// Error is always nil.
	w.buf.Write(p)
	n := len(p)
	for {
		line, err := w.buf.ReadString('\n')
		if err != nil {
			// Did not have a whole line to read, rebuffer the unconsumed data.
			// Error is always nil.
			w.buf.Write([]byte(line))
			return n, nil
		}
		// Write non-empty lines from the buffer.
		if !isBlank(line) {
			if _, err := w.output.Write([]byte(line)); err != nil {
				return n, err
			}
		}
	}
}
