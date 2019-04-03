package main

import (
	"fmt"
	"strings"
)

// A Function definition.
type Function struct {
	Name       string // C name of the function
	GoName     string // Go name of the function with the API prefix stripped
	Parameters []Parameter
	Return     Type
}

// IsImplementedForSyscall reports whether the function is implemented for syscall or not.
func (f Function) IsImplementedForSyscall() bool {
	// TODO: Use syscall.Syscall18 when Go 1.12 is the minimum supported version.
	if len(f.Parameters) > 15 {
		return false
	}
	return true
}

// Syscall returns a syscall expression for Windows.
func (f Function) Syscall() string {
	var ps []string
	for _, p := range f.Parameters {
		ps = append(ps, p.Type.ConvertGoToUintptr(p.GoName()))
	}
	for len(ps) == 0 || len(ps)%3 != 0 {
		ps = append(ps, "0")
	}

	post := ""
	if len(ps) > 3 {
		post = fmt.Sprintf("%d", len(ps))
	}

	return fmt.Sprintf("syscall.Syscall%s(gp%s, %d, %s)", post, f.GoName, len(f.Parameters), strings.Join(ps, ", "))
}

// A Parameter to a Function.
type Parameter struct {
	Name string
	Type Type
}

// CName returns a C-safe parameter name.
func (p Parameter) CName() string {
	return renameIfReservedCWord(p.Name)
}

// GoName returns a Go-safe parameter name.
func (p Parameter) GoName() string {
	return renameIfReservedGoWord(p.Name)
}

func renameIfReservedCWord(word string) string {
	switch word {
	case "near", "far":
		return fmt.Sprintf("x%s", word)
	}
	return word
}

func renameIfReservedGoWord(word string) string {
	switch word {
	case "func", "type", "struct", "range", "map", "string":
		return fmt.Sprintf("x%s", word)
	}
	return word
}
