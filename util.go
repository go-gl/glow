package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type ParamLenType int

const (
	ParamLenTypeUnknown ParamLenType = iota
	ParamLenTypeParamRef
	ParamLenTypeValue
	ParamLenTypeCompSize
)

type ParamLen struct {
	Type     ParamLenType
	ParamRef string
	Value    int
	Params   string
}

func TrimGLCmdPrefix(str string) string {
	if strings.HasPrefix(str, "gl") {
		return strings.TrimPrefix(str, "gl")
	}
	if strings.HasPrefix(str, "glx") {
		return strings.TrimPrefix(str, "glx")
	}
	if strings.HasPrefix(str, "wgl") {
		return strings.TrimPrefix(str, "wgl")
	}
	return str
}

func TrimGLEnumPrefix(str string) string {
	t := str
	p := ""
	if strings.HasPrefix(str, "GL_") {
		t = strings.TrimPrefix(t, "GL_")
		p = "GL_"
	} else if strings.HasPrefix(str, "GLX_") {
		t = strings.TrimPrefix(str, "GLX_")
		p = "GLX_"
	} else if strings.HasPrefix(str, "WGL_") {
		t = strings.TrimPrefix(str, "WGL_")
		p = "WGL_"
	}
	if strings.IndexAny(t, "0123456789") == 0 {
		return p + t
	}
	return t
}

func ParseLenString(lenStr string) ParamLen {
	if strings.HasPrefix(lenStr, "COMPSIZE") {
		p := strings.TrimSuffix(strings.TrimPrefix(lenStr, "COMPSIZE("), ")")
		return ParamLen{Type: ParamLenTypeCompSize, Params: p}
	}
	n, err := strconv.ParseInt(lenStr, 10, 32)
	if err == nil {
		return ParamLen{Type: ParamLenTypeValue, Value: (int)(n)}
	}
	return ParamLen{Type: ParamLenTypeParamRef, ParamRef: lenStr}
}

// Prevent name clashes.
func RenameIfReservedCWord(word string) string {
	switch word {
	case "near", "far":
		return fmt.Sprintf("%s", word)
	}
	return word
}

// Prevent name clashes.
func RenameIfReservedGoWord(word string) string {
	switch word {
	case "func", "type", "struct", "range", "map", "string":
		return fmt.Sprintf("gl%s", word)
	}
	return word
}

// Converts strings with underscores to Go-like names. e.g.: bla_blub_foo -> BlaBlubFoo
func CamelCase(n string) string {
	prev := '_'
	return strings.Map(
		func(r rune) rune {
			if r == '_' {
				prev = r
				return -1
			}
			if prev == '_' {
				prev = r
				return unicode.ToTitle(r)
			}
			prev = r
			return r
		},
		n)
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
