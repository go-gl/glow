package main

import "fmt"

// A Function definition.
type Function struct {
	Name       string // C name of the function
	GoName     string // Go name of the function with the API prefix stripped
	Parameters []Parameter
	Return     Type
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
