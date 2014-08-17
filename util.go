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
//
// Comment-based annotations are accepted to define sections of code that
// should have their blank lines kept intact, like so:
//
//  //
//  //glow:keepspace
//  //
//  // Hello World!
//  //
//  //glow:rmspace
//  //
//
// The writer would produce output like:
//  //
//  // Hello World!
//  //
//
type BlankLineStrippingWriter struct {
	output    io.Writer
	buf       *bytes.Buffer
	stripping bool
}

// NewBlankLineStrippingWriter creates a new BlankLineStrippingWriter.
func NewBlankLineStrippingWriter(wrapped io.Writer) *BlankLineStrippingWriter {
	return &BlankLineStrippingWriter{
		output:    wrapped,
		buf:       new(bytes.Buffer),
		stripping: true,
	}
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
func (w *BlankLineStrippingWriter) Write(p []byte) (int, error) {
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

		// Enable/disable blank line stripping based on comment-based
		// annotations.
		cleanLine := strings.TrimSpace(line)
		if cleanLine == "//glow:keepspace" {
			w.stripping = false
			continue
		} else if cleanLine == "//glow:rmspace" {
			w.stripping = true
			continue
		}

		// Write non-empty lines from the buffer.
		if !w.stripping || !isBlank(line) {
			if _, err := w.output.Write([]byte(line)); err != nil {
				return n, err
			}
		}
	}
}
