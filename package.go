package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// A Package holds the typedef, function, and enum definitions for a Go package.
type Package struct {
	Name    string
	API     string
	Version Version
	Profile string
	TmplDir string

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

// Comment returns the comment explaining the function.
func (f *PackageFunction) Comment() string {
	var lines []string
	if f.Doc != "" {
		lines = append(lines, "// " + f.Doc)
	}

	// Adds explanations about C types that are unsafe.Pointer in Go world.
	// See also https://github.com/go-gl/gl/issues/113.
	for _, p := range f.Function.Parameters {
		if t := p.Type; t.GoType() == "unsafe.Pointer" && t.Name != "void" && t.Name != "GLvoid" {
			lines = append(lines, fmt.Sprintf("// Parameter %s has type %s.", p.Name, t.GoCType()))
		}
	}

	if r := f.Function.Return; r.GoType() == "unsafe.Pointer" && r.Name != "void" && r.Name != "GLvoid" {
		lines = append(lines, fmt.Sprintf("// Return value has type %s.", r.GoCType()))
	}

	return strings.Join(lines, "\n")
}

// UniqueName returns a globally unique Go-compatible name for this package.
func (pkg *Package) UniqueName() string {
	version := strings.Replace(pkg.Version.String(), ".", "", -1)
	return fmt.Sprintf("%s%s%s", pkg.API, pkg.Profile, version)
}

// GeneratePackage writes a Go package to specified directory.
func (pkg *Package) GeneratePackage(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := pkg.generateFile("package", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("package_notwindows", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("package_windows", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("conversions", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("conversions_notwindows", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("conversions_windows", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("procaddr_notwindows", dir); err != nil {
		return err
	}
	if err := pkg.generateFile("procaddr_windows", dir); err != nil {
		return err
	}
	if pkg.HasDebugCallbackFeature() {
		if err := pkg.generateFile("debug_notwindows", dir); err != nil {
			return err
		}
		if err := pkg.generateFile("debug_windows", dir); err != nil {
			return err
		}
	}

	// gofmt the generated .go files.
	if err := exec.Command("gofmt", "-w", dir).Run(); err != nil {
		return fmt.Errorf("gofmt error: %v", err)
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

	tmpl := template.Must(template.New(file + ".tmpl").Funcs(fns).ParseFiles(filepath.Join(pkg.TmplDir, file+".tmpl")))

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

// importPathToDir resolves the absolute path from importPath.
// There needs to be a valid Go package inside that import path.
// It calls log.Fatalln if it fails.
func importPathToDir(importPath string) string {
	pkgs, err := packages.Load(nil, importPath)
	if err != nil {
		log.Fatalln(err)
	}
	return filepath.Dir(pkgs[0].GoFiles[0])
}
