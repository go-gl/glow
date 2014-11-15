package main

import (
	"fmt"
	"strings"
)

// A Type describes the C and Go type of a function parameter or return value.
type Type struct {
	Name         string // Name of the type without modifiers
	PointerLevel int    // Number of levels of declared indirection to the type
	CDefinition  string // Raw C definition
}

// A Typedef describes a C typedef statement.
type Typedef struct {
	Name        string // Name of the defined type (or included types)
	CDefinition string // Raw C definition
}

func (t Type) String() string {
	return fmt.Sprintf("%s%s [%s]", t.Name, t.pointers(), t.CDefinition)
}

// IsVoid indicates whether this type is the void pseudo-type.
func (t Type) IsVoid() bool {
	return (t.Name == "void" || t.Name == "GLvoid") && t.PointerLevel == 0
}

// IsDebugProc indicates whether this type is a debug callback function pointer.
func (t Type) IsDebugProc() bool {
	return t.Name == "GLDEBUGPROC" || t.Name == "GLDEBUGPROCARB" || t.Name == "GLDEBUGPROCKHR"
}

// CType returns the C definition of the type.
func (t Type) CType() string {
	return t.CDefinition
}

// GoCType returns the Go definition of the C type.
func (t Type) GoCType() string {
	if strings.HasPrefix(t.Name, "struct ") {
		return t.pointers() + "C.struct_" + strings.TrimPrefix(t.Name, "struct ")
	}
	return t.pointers() + "C." + t.Name
}

// GoType returns the Go definition of the type.
func (t Type) GoType() string {
	switch t.Name {
	case "GLbyte":
		return t.pointers() + "int8"
	case "GLubyte":
		return t.pointers() + "uint8"
	case "GLshort":
		return t.pointers() + "int16"
	case "GLushort":
		return t.pointers() + "uint16"
	case "GLint":
		return t.pointers() + "int32"
	case "GLuint":
		return t.pointers() + "uint32"
	case "GLint64", "GLint64EXT":
		return t.pointers() + "int64"
	case "GLuint64", "GLuint64EXT":
		return t.pointers() + "uint64"
	case "GLfloat", "GLclampf":
		return t.pointers() + "float32"
	case "GLdouble", "GLclampd":
		return t.pointers() + "float64"
	case "GLclampx":
		return t.pointers() + "int32"
	case "GLsizei":
		return t.pointers() + "int32"
	case "GLfixed":
		return t.pointers() + "int32"
	case "GLchar", "GLcharARB":
    // Chosen (vs. byte) for compatibility where GL uses GLubyte for strings (e.g., glGetString)
		return t.pointers() + "uint8"
	case "GLboolean":
		return t.pointers() + "bool"
	case "GLenum", "GLbitfield":
		return t.pointers() + "uint32"
	case "GLhalf", "GLhalfNV":
		// Go has no 16-bit floating point type
		return t.pointers() + "uint16"
	case "void", "GLvoid":
		// Type void* could map to either uintptr or unsafe.Pointer but we use the
		// latter because pointers passed from Go to C need to use Go pointer types
		// for the sake of the GC
		if t.PointerLevel == 1 {
			return "unsafe.Pointer"
		} else if t.PointerLevel == 2 {
			return "*unsafe.Pointer"
		}
	case "GLintptr", "GLintptrARB":
		// Type ptrdiff_t should map to an architecture-width Go int
		return t.pointers() + "int"
	case "GLsizeiptr", "GLsizeiptrARB":
		// Same as GLintptr
		return t.pointers() + "int"
	case "GLhandleARB", "GLeglImagesOES", "GLvdpauSurfaceNV":
		// OpenGL pointer types should fit in a pointer-width Go type. No need to
		// use unsafe.Pointer because there is no need for the Go GC to understand
		// that these are pointer types. Moreover on most platforms GLhandleARB is
		// an integer type.
		return t.pointers() + "uintptr"
	case "GLsync":
		return t.pointers() + "unsafe.Pointer"
	case "GLDEBUGPROC", "GLDEBUGPROCARB", "GLDEBUGPROCKHR":
		// Special case mapping to the type defined in debug.tmpl
		return "DebugProc"
	}
	return t.GoCType()
}

// ConvertGoToC returns an expression that converts a variable from the Go type to the C type.
func (t Type) ConvertGoToC(name string) string {
	switch t.Name {
	case "GLboolean":
		if t.PointerLevel == 0 {
			return fmt.Sprintf("(C.GLboolean)(boolToInt(%s))", name)
		}
	case "void", "GLvoid":
		return name
	case "GLDEBUGPROC", "GLDEBUGPROCARB", "GLDEBUGPROCKHR":
		return fmt.Sprintf("(C.%s)(unsafe.Pointer(&%s))", t.Name, name)
	}
	if t.PointerLevel >= 1 {
		return fmt.Sprintf("(%s)(unsafe.Pointer(%s))", t.GoCType(), name)
	}
	return fmt.Sprintf("(%s)(%s)", t.GoCType(), name)
}

// ConvertCToGo converts from the C type to the Go type.
func (t Type) ConvertCToGo(name string) string {
	if t.Name == "GLboolean" {
		return fmt.Sprintf("%s == TRUE", name)
	}
	return fmt.Sprintf("(%s)(%s)", t.GoType(), name)
}

func (t Type) pointers() string {
	return strings.Repeat("*", t.PointerLevel)
}
