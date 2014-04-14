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
	TypeDefs  []TypeDef
	Enums     Enums
	Functions Functions
}

type Packages []*Package

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
	}
	tmpl := template.Must(template.New("commands.tmpl").Funcs(fns).ParseFiles("commands.tmpl"))
	return tmpl.Execute(NewBlankLineStrippingWriter(w), p)
}

func (p *Package) GeneratePackage() error {
	fmt.Println("Generating package", p.Name, p.Version)
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

func (ps Packages) GeneratePackages() error {
	for _, p := range ps {
		err := p.GeneratePackage()
		if err != nil {
			return err
		}
	}
	return nil
}
