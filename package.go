package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Package struct {
	Name      string
	Api       string
	Version   Version
	Typedefs  []*Typedef
	Enums     Enums
	Functions Functions
}

const docURLFmt = "https://www.opengl.org/sdk/docs/man%d/html/%s%s.xhtml"

// TODO Map documentation URL by version number
// TODO Strip suffixes to construct URLs

func (p *Package) doc(fnName string) string {
	return fmt.Sprintf(docURLFmt, p.Version.Major, p.Api, fnName)
}

func (p *Package) writeEnums(dir string) error {
	w, err := os.Create(filepath.Join(dir, "enums.go"))
	if err != nil {
		return err
	}
	defer w.Close()
	tmpl := template.Must(template.ParseFiles("enums.tmpl"))
	return tmpl.Execute(NewBlankLineStrippingWriter(w), p.Enums.Sort())
}

func (p *Package) writeCommands(dir string) error {
	w, err := os.Create(filepath.Join(dir, "commands.go"))
	if err != nil {
		return err
	}
	defer w.Close()
	fns := template.FuncMap{
		"replace": strings.Replace,
		"toUpper": strings.ToUpper,
		"doc":     p.doc,
	}
	tmpl := template.Must(template.New("commands.tmpl").Funcs(fns).ParseFiles("commands.tmpl"))
	return tmpl.Execute(NewBlankLineStrippingWriter(w), p)
}

func (p *Package) GeneratePackage() error {
	dir := filepath.Join(p.Api, p.Version.String(), p.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := p.writeEnums(dir); err != nil {
		return err
	}
	if err := p.writeCommands(dir); err != nil {
		return err
	}
	return nil
}
