package main

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// A Package holds the typedef, function, and enum definitions for a Go package.
type Package struct {
	Name      string
	Api       string
	Version   Version
	Typedefs  []*Typedef
	Enums     map[string]Enum
	Functions map[string]PackageFunction
}

type PackageFunction struct {
	Function   Function
	Required   bool
	Extensions []string
}

// GeneratePackage writes a Go package file.
func (pkg *Package) GeneratePackage() error {
	dir := filepath.Join(pkg.Api, pkg.Version.String(), pkg.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(dir, pkg.Name+".go"))
	if err != nil {
		return err
	}
	defer out.Close()

	fns := template.FuncMap{
		"replace": strings.Replace,
		"toUpper": strings.ToUpper,
	}
	tmpl := template.Must(template.New("package.tmpl").Funcs(fns).ParseFiles("package.tmpl"))

	return tmpl.Execute(NewBlankLineStrippingWriter(out), pkg)
}
