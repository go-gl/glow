package main

import (
	"encoding/xml"
	"os"
	"strings"
)

type xmlDoc struct {
	Purpose string   `xml:"refnamediv>refpurpose"`
	Names   []string `xml:"refnamediv>refname"`
}

// Documentation is a map from function name to documentation string.
type Documentation map[string]string

func readDocFile(file string) (*xmlDoc, error) {
	var xmlDoc xmlDoc

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := xml.NewDecoder(f)
	decoder.Strict = false

	err = decoder.Decode(&xmlDoc)
	if err != nil {
		return nil, err
	}

	return &xmlDoc, nil
}

// AddDocs adds function documentation to the specified package.
func (d Documentation) AddDocs(pkg *Package) {
	for _, fn := range pkg.Functions {
		doc, ok := d[fn.Name]
		if ok {
			// Let the template handle line-wrapping if it chooses
			fn.Doc = strings.Replace(doc, "\n", " ", -1)
		}
	}
}

// NewDocumentation parses XML documentation files.
func NewDocumentation(files []string) (Documentation, error) {
	documentation := make(Documentation)
	for _, file := range files {
		xmlDoc, err := readDocFile(file)
		if err != nil {
			return nil, err
		}
		for _, name := range xmlDoc.Names {
			documentation[name] = xmlDoc.Purpose
		}
	}
	return documentation, nil
}
