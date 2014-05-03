package main

import (
	"fmt"
	"strings"
)

type Type struct {
	Name         string // Name of the type without modifiers
	PointerLevel int    // Number of levels of declared indirection to the type
	CDefinition  string // Raw C definition
}

type Typedef struct {
	Name        string // Name of the defined type (or included types)
	CDefinition string // Raw C definition
}

func (t Type) String() string {
	return fmt.Sprintf("%s%s [%s]", t.Name, t.pointers(), t.CDefinition)
}

func (t Type) pointers() string {
	return strings.Repeat("*", t.PointerLevel)
}

func (t Type) IsVoid() bool {
	return (t.Name == "void" || t.Name == "GLvoid") && t.PointerLevel == 0
}

// CType returns the C definition of the type.
func (t Type) CType() string {
	return t.CDefinition
}

// GoType returns the Go definition of the type.
func (t Type) GoType() string {
	switch t.Name {
	case "GLenum":
		return t.pointers() + "glt.Enum"
	case "GLbitfield":
		return t.pointers() + "glt.Bitfield"
	case "GLboolean":
		if t.PointerLevel == 0 {
			return "bool"
		}
		return t.pointers() + "byte"
	case "GLint":
		return t.pointers() + "int32"
	case "GLuint":
		return t.pointers() + "uint32"
	case "GLint64", "GLint64EXT":
		return t.pointers() + "int64"
	case "GLuint64", "GLuint64EXT":
		return t.pointers() + "uint64"
	case "GLclampf", "GLfloat":
		return t.pointers() + "float32"
	case "GLclampd", "GLdouble":
		return t.pointers() + "float64"
	case "GLclampx":
		return t.pointers() + "int32"
	case "GLsizei":
		return t.pointers() + "int32"
	case "GLbyte":
		return t.pointers() + "int8"
	case "GLfixed":
		return t.pointers() + "int32"
	case "void", "GLvoid":
		if t.PointerLevel == 1 {
			return "glt.Pointer"
		} else if t.PointerLevel == 2 {
			return "*glt.Pointer"
		}
		return ""
	case "GLintptr", "GLintptrARB":
		if t.PointerLevel == 0 {
			return "int"
		}
		return t.pointers() + "int64"
	case "GLsizeiptrARB", "GLsizeiptr":
		if t.PointerLevel == 0 {
			return "int"
		}
		return t.pointers() + "int64"
	case "GLcharARB", "GLchar":
		return t.pointers() + "int8"
	case "GLubyte":
		return t.pointers() + "uint8"
	case "GLshort":
		return t.pointers() + "int16"
	case "GLushort":
		return t.pointers() + "uint16"
	case "GLhandleARB":
		return t.pointers() + "glt.Pointer"
	case "GLhalfNV":
		return t.pointers() + "uint16"
	case "GLeglImageOES":
		return t.pointers() + "glt.Pointer"
	case "GLvdpauSurfaceARB":
		return t.pointers() + "glt.Pointer"
	case "GLsync":
		return t.pointers() + "glt.Sync"
	case "void **":
		return "*glt.Pointer"
	case "const void *const*":
		return "*glt.Pointer"
	case "GLDEBUGPROC":
		return "glt.DebugProc"
	}
	return t.pointers() + "C." + t.Name
}

// ConvertGoToC converts from the Go type to the C type.
func (t Type) ConvertGoToC() string {
	switch t.Name {
	case "GLboolean":
		if t.PointerLevel == 0 {
			return "GoBoolean"
		}
	case "void", "GLvoid":
		if t.PointerLevel == 1 {
			return "unsafe.Pointer"
		} else if t.PointerLevel == 2 {
			return "cgoPtr1"
		}
	case "GLchar":
		if t.PointerLevel == 2 {
			return "cgoChar2"
		}
	}
	return fmt.Sprintf("(%sC.%s)", t.pointers(), t.Name)
}

// ConvertCToGo converts from the C type to the Go type.
func (t Type) ConvertCToGo() string {
	switch t.Name {
	case "GLboolean":
		if t.PointerLevel == 0 {
			return "GLBoolean"
		}
	case "void", "GLvoid":
		if t.PointerLevel > 0 {
			return "glt.Pointer"
		}
	case "GLintptr", "GLintptrARB":
		if t.PointerLevel == 0 {
			return "int"
		}
	case "GLuint", "GLuintARB":
		if t.PointerLevel == 0 {
			return "uint32"
		}
	case "GLenum":
		if t.PointerLevel == 0 {
			return "glt.Enum"
		}
	case "GLubyte":
		return "(" + t.pointers() + "byte)"
	case "GLint":
		return t.pointers() + "int32"
	case "GLsizeiptrARB", "GLsizeiptr":
		if t.PointerLevel == 0 {
			return "int"
		}
	case "GLsync":
		return "glt.Sync"
	}
	return fmt.Sprintf("<unknown type:%sC.%s>", t.pointers(), t.Name)
}
