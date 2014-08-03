package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// A Package holds the typedef, function, and enum definitions for a Go package.
type Package struct {
	Name    string
	API     string
	Version Version
	Profile string

	Typedefs  []*Typedef
	Enums     map[string]*Enum
	Functions map[string]*PackageFunction
}

// A PackageFunction is a package-specific Function wrapper.
type PackageFunction struct {
	Function
	Required bool
	Doc      string
}

// Dir returns the directory to which the Go package files are written.
func (pkg *Package) Dir() string {
	apiPrefix := pkg.API
	if pkg.Profile != "" {
		apiPrefix = pkg.API + "-" + pkg.Profile
	}
	return filepath.Join(apiPrefix, pkg.Version.String(), pkg.Name)
}

// UniqueName returns a globally unique Go-compatible name for thie package.
func (pkg *Package) UniqueName() string {
	version := strings.Replace(pkg.Version.String(), ".", "", -1)
	return fmt.Sprintf("%s%s%s", pkg.API, pkg.Profile, version)
}

// GeneratePackage writes a Go package file.
func (pkg *Package) GeneratePackage() error {
	dir := pkg.Dir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := pkg.generateFile("package", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("conversions", dir); err != nil {
		return err
	}
	if pkg.HasDebugCallbackFeature() {
		if err := pkg.generateFile("debug", dir); err != nil {
			return err
		}
	}

	return nil
}

func (pkg *Package) generateFile(file, dir string) error {
	out, err := os.Create(filepath.Join(dir, file+".go"))
	if err != nil {
		return err
	}
	defer out.Close()

	fns := template.FuncMap{
		"replace": strings.Replace,
		"toUpper": strings.ToUpper,
	}
	tmpl := template.Must(template.New(file + ".tmpl").Funcs(fns).ParseFiles(file + ".tmpl"))

	return tmpl.Execute(NewBlankLineStrippingWriter(out), pkg)
}

// HasDebugCallbackFeature returns whether this package exposes the ability to
// set a debug callback. Used to determine whether to include the necessary
// GL-specific callback code.
func (pkg *Package) HasDebugCallbackFeature() bool {
	for _, fn := range pkg.Functions {
		for _, param := range fn.Parameters {
			if param.Type.IsDebugProc() {
				return true
			}
		}
	}
	return false
}

// HasRequiredFunctions returns true if at least one function in this package
// is required.
func (pkg *Package) HasRequiredFunctions() bool {
	for _, fn := range pkg.Functions {
		if fn.Required {
			return true
		}
	}
	return false
}
