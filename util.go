package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func TrimGLCmdPrefix(cmdName string) string {
	if strings.HasPrefix(cmdName, "gl") {
		return strings.TrimPrefix(cmdName, "gl")
	}
	if strings.HasPrefix(cmdName, "glx") {
		return strings.TrimPrefix(cmdName, "glx")
	}
	if strings.HasPrefix(cmdName, "wgl") {
		return strings.TrimPrefix(cmdName, "wgl")
	}
	return cmdName
}

func TrimGLEnumPrefix(enumName string) string {
	trimmed := enumName
	prefix := ""
	if strings.HasPrefix(enumName, "GL_") {
		trimmed = strings.TrimPrefix(enumName, "GL_")
		prefix = "GL_"
	} else if strings.HasPrefix(enumName, "GLX_") {
		trimmed = strings.TrimPrefix(enumName, "GLX_")
		prefix = "GLX_"
	} else if strings.HasPrefix(enumName, "WGL_") {
		trimmed = strings.TrimPrefix(enumName, "WGL_")
		prefix = "WGL_"
	}
	if strings.IndexAny(trimmed, "0123456789") == 0 {
		return prefix + trimmed
	}
	return trimmed
}

func RenameIfReservedCWord(word string) string {
	switch word {
	case "near", "far":
		return fmt.Sprintf("%s", word)
	}
	return word
}

func RenameIfReservedGoWord(word string) string {
	switch word {
	case "func", "type", "struct", "range", "map", "string":
		return fmt.Sprintf("gl%s", word)
	}
	return word
}

// Writer that removes whitespace- or comment-only lines delimited by \n
// A necessary evil to work around how text/template handles whitespace
type BlankLineStrippingWriter struct {
	output io.Writer
	buf    *bytes.Buffer
}

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

func (w BlankLineStrippingWriter) Write(p []byte) (n int, err error) {
	// Buffer the current write
	nn, err := w.buf.Write(p)
	if nn != len(p) || err != nil {
		return 0, err
	}
	// Write non-empty lines from the buffer
	for {
		line, err := w.buf.ReadString('\n')
		switch err {
		case nil:
			if !isBlank(line) {
				nn, e := w.output.Write([]byte(line))
				if nn != len(line) || e != nil {
					return n, err
				}
			}
			n += len(line)
		case io.EOF:
			// Did not have a whole line to read, rebuffer the unconsumed data
			w.buf.Write([]byte(line))
			return 0, nil
		default:
			return 0, err
		}
	}
	return n, err
}
