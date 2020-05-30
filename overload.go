package main

import (
	"encoding/xml"
	"os"
)

type xmlOverloads struct {
	Overloads []xmlOverload `xml:"overload"`
}

type xmlOverload struct {
	Name         string `xml:"name,attr"`
	OverloadName string `xml:"overloadName,attr"`

	ParameterChanges []xmlParameterChange `xml:"parameterChanges>change"`
}

type xmlParameterChange struct {
	// Index is the zero-based index of the parameter list.
	Index int `xml:"index,attr"`
	// Name describes a change of the parameter name.
	Name *xmlNameChange `xml:"name"`
	// Type describes a change of the parameter type.
	Type *xmlTypeChange `xml:"type"`
}

type xmlNameChange struct {
	Value string `xml:"value,attr"`
}

type xmlTypeChange struct {
	Signature string `xml:"signature,attr"`
}

func readOverloadFile(file string) (xmlOverloads, error) {
	var overloads xmlOverloads

	_, err := os.Stat(file)
	if err != nil {
		return overloads, nil
	}

	f, err := os.Open(file)
	if err != nil {
		return overloads, err
	}
	defer f.Close()

	err = xml.NewDecoder(f).Decode(&overloads)
	return overloads, err
}
