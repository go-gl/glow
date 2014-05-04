package main

// A Function.
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
	return RenameIfReservedCWord(p.Name)
}

// GoName returns a Go-safe parameter name.
func (p Parameter) GoName() string {
	return RenameIfReservedGoWord(p.Name)
}
