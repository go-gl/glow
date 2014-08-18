package main

import (
	"fmt"
	"go/build"
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

	SpecRev string
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
	filePath := pkgPath(file + ".tmpl")
	tmpl := template.Must(template.New(file + ".tmpl").Funcs(fns).ParseFiles(filePath))

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

// Filter removes any enums, or functions found in this package that are not
// listed in the given lookup maps. If either of the maps has a length of zero,
// filtering does not occur for that type (e.g. all functions are left intact).
func (pkg *Package) Filter(enums, functions map[string]bool) {
	if len(enums) > 0 {
		// Remove any enum not listed in the enums lookup map.
		for name := range pkg.Enums {
			_, valid := enums[name]
			if !valid {
				delete(pkg.Enums, name)
			}
		}
	}

	if len(functions) > 0 {
		// Remove any function not listed in the functions lookup map.
		for name := range pkg.Functions {
			_, valid := functions[name]
			if !valid {
				delete(pkg.Functions, name)
			}
		}
	}
}

// This package's directory (relative to $GOPATH).
const pkgDir = "src/github.com/go-gl/glow"

// Cached absolute path to this package's directory (determined in the pkgPath
// function below).
var absPkgDir string

// pkgPath resolves a path relative to the package directory (see above).
func pkgPath(relPath string) string {
	if len(absPkgDir) == 0 {
		// Find absolute path to this package's directory.
		for _, path := range filepath.SplitList(build.Default.GOPATH) {
			path = filepath.Join(path, pkgDir)
			if _, err := os.Stat(path); err == nil {
				absPkgDir = path
				break
			}
		}
	}
	return filepath.Join(absPkgDir, relPath)
}
