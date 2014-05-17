package main

// An enumerated value.
type Enum struct {
	Name   string // Raw specification name
	GoName string // Go name with the API prefix stripped
	Value  string // Raw specification value
}
