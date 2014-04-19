package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type SpecRegistry struct {
	XMLName  xml.Name        `xml:"registry"`
	Comment  string          `xml:"comment"`
	Types    []SpecType      `xml:"types>type"`
	Groups   []SpecGroup     `xml:"groups>group"`
	Enums    []SpecEnumToken `xml:"enums"`
	Commands []SpecCommand   `xml:"commands>command"`
	Features []SpecFeature   `xml:"feature"`
	// TODO: extensions
}

type SpecType struct {
	Name     string `xml:"name,attr"`
	Comment  string `xml:"comment,attr"`
	Requires string `xml:"requires,attr"`
	Api      string `xml:"api,attr"`
	Inner    []byte `xml:",innerxml"`
}

type SpecGroup struct {
	Name    string      `xml:"name,attr"`
	Comment string      `xml:"comment,attr"`
	Enums   []SpecGEnum `xml:"enum"`
}

type SpecGEnum struct {
	Name string `xml:"name,attr"`
}

type SpecEnumToken struct {
	Namespace string      `xml:"namespace,attr"`
	Group     string      `xml:"group,attr"`
	Type      string      `xml:"type,attr"`
	Comment   string      `xml:"comment,attr"`
	Enums     []SpecTEnum `xml:"enum"`
	// TODO: vendor, start, end attr
}

type SpecTEnum struct {
	Value string `xml:"value,attr"`
	Name  string `xml:"name,attr"`
}

type SpecCommand struct {
	Proto  SpecProto   `xml:"proto"`
	Params []SpecParam `xml:"param"`
}

type SpecSignature []byte

type SpecProto struct {
	Inner SpecSignature `xml:",innerxml"`
}

type SpecParam struct {
	Group string        `xml:"group,attr"`
	Len   string        `xml:"len,attr"`
	Inner SpecSignature `xml:",innerxml"`
}

type SpecFeature struct {
	Api      string        `xml:"api,attr"`
	Name     string        `xml:"name,attr"`
	Number   string        `xml:"number,attr"`
	Requires []SpecRequire `xml:"require"`
	Removes  []SpecRemove  `xml:"remove"`
}

type SpecRequire struct {
	Comment  string           `xml:"comment,attr"`
	Enums    []SpecEnumRef    `xml:"enum"`
	Commands []SpecCommandRef `xml:"command"`
}

type SpecRemove struct {
	Comment  string           `xml:"comment,attr"`
	Enums    []SpecEnumRef    `xml:"enum"`
	Commands []SpecCommandRef `xml:"command"`
}

type SpecEnumRef struct {
	Name string `xml:"name,attr"`
}

type SpecCommandRef struct {
	Name string `xml:"name,attr"`
}

func (st *SpecType) Parse() (TypeDef, error) {
	typed := TypeDef{Name: st.Name, Comment: st.Comment, Api: st.Api, CDefinition: ""}
	readName := false
	decoder := xml.NewDecoder(bytes.NewBuffer(st.Inner))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return typed, err
		}
		switch t := token.(type) {
		case xml.CharData:
			typed.CDefinition += (string)(t)
			if readName {
				typed.Name = (string)(t)
			}
		case xml.StartElement:
			if t.Name.Local == "name" {
				readName = true
			} else if t.Name.Local == "apientry" {
				typed.CDefinition += "APIENTRY"
			} else {
				return typed, fmt.Errorf("Wrong start element: %s", t.Name.Local)
			}
		case xml.EndElement:
			if t.Name.Local == "name" {
				readName = false
			} else if t.Name.Local == "apientry" {
			} else {
				return typed, fmt.Errorf("Wrong start element: %s", t.Name.Local)
			}
		}
	}
	return typed, nil
}

func (r SpecRegistry) ParseTypedefs() ([]TypeDef, error) {
	tdefs := make([]TypeDef, 0, len(r.Types))
	for _, s := range r.Types {
		td, err := s.Parse()
		if err != nil {
			return nil, err
		}
		tdefs = append(tdefs, td)
	}
	return tdefs, nil
}

func (si SpecSignature) Parse() (string, Type, error) {
	name := ""
	ctype := Type{}
	readName := false
	readType := false
	first := true
	decoder := xml.NewDecoder(bytes.NewBuffer(si))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return name, ctype, err
		}
		switch t := token.(type) {
		case xml.CharData:
			s := strings.Trim((string)(t), " ")
			if readName {
				name = TrimGLCmdPrefix(s)
			} else if readType {
				ctype.Name = s
			} else if s == "" {
				// skip
			} else if s == "void" {
				ctype.Name = "void"
			} else if s == "void *" {
				ctype.Name = "void"
				ctype.PointerLevel = 1
			} else if s == "const void *" {
				ctype.Name = "void"
				ctype.IsConst = true
				ctype.PointerLevel = 1
			} else if s == "*" {
				ctype.PointerLevel = 1
			} else if s == "**" {
				ctype.PointerLevel = 2
			} else if s == "*const*" {
				ctype.PointerLevel = 2
			} else if s == "const" {
				ctype.IsConst = true
			} else if first {
				ctype.Name = s
				first = false
			} else {
				return name, ctype, fmt.Errorf("Unknown %s", s)
			}
		case xml.StartElement:
			if t.Name.Local == "ptype" {
				readType = true
			} else if t.Name.Local == "name" {
				readName = true
			} else {
				return name, ctype, fmt.Errorf("Wrong start element: %s", t.Name.Local)
			}
		case xml.EndElement:
			if t.Name.Local == "ptype" {
				readType = false
			} else if t.Name.Local == "name" {
				readName = false
			} else {
				return name, ctype, fmt.Errorf("Wrong end element: %s", t.Name.Local)
			}
		}
	}
	return name, ctype, nil
}

func readSpecFile(file string) (*SpecRegistry, error) {
	var reg SpecRegistry
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	err = d.Decode(&reg)
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func commandsToFunctions(commands []SpecCommand) Functions {
	functions := make(Functions)
	for _, c := range commands {
		cname, ct, err := c.Proto.Inner.Parse()
		if err != nil {
			fmt.Printf("Unable to parse proto signature '%s': %s\n", string(c.Proto.Inner), err)
		} else {
			parameters := make([]Parameter, 0, 4)
			for _, p := range c.Params {
				pname, pt, err := p.Inner.Parse()
				if err != nil {
					fmt.Printf("Unable to parse parameter signature '%s' of function '%s': %s\n", (string)(p.Inner), cname, err)
				} else {
					parameters = append(parameters, Parameter{Name: pname, Type: pt})
				}
			}
			functions[cname] = &Function{Name: cname, Parameters: parameters, Return: ct}
		}
	}
	return functions
}

func findEnum(enumName string, est []SpecEnumToken) (string, string) {
	for _, es := range est {
		for _, e := range es.Enums {
			if e.Name == enumName {
				return e.Value, es.Group
			}
		}
	}
	return "", ""
}

func addEnums(ps Packages, api string, ver Version, enumNames []SpecEnumRef, et []SpecEnumToken) {
	fmt.Println("Adding enums from version", api, ver, "to")
	for _, pc := range ps {
		if pc.Api != api {
			continue
		}
		if pc.Version.Compare(ver) < 0 {
			continue
		}
		fmt.Println(" package", pc.Api, pc.Version)
		for _, en := range enumNames {
			val, grp := findEnum(en.Name, et)
			if val == "" {
				fmt.Println("Not found:", en.Name)
			}
			pc.Enums[en.Name] = &Enum{Name: TrimGLEnumPrefix(en.Name), Value: val, Group: grp}
		}
	}
}

func removeEnums(ps Packages, api string, ver Version, enumNames []SpecEnumRef) {
	fmt.Println("Removing enums from version", api, ver, "to")
	for _, pc := range ps {
		if pc.Api != api {
			continue
		}
		if pc.Version.Compare(ver) < 0 {
			continue
		}
		fmt.Println(" package", pc.Api, pc.Version)
		for _, en := range enumNames {
			if _, ok := pc.Enums[en.Name]; ok {
				delete(pc.Enums, en.Name)
			}
		}
	}
}

func addCommands(pkgs Packages, api string, ver Version, cmdNames []SpecCommandRef, functions Functions) {
	for _, pkg := range pkgs {
		if pkg.Api != api {
			continue
		}
		if pkg.Version.Compare(ver) < 0 {
			continue
		}
		for _, cmd := range cmdNames {
			fnName := TrimGLCmdPrefix(cmd.Name)
			fn, ok := functions[fnName]
			if !ok {
				log.Fatal("Function not found", fnName)
			}
			pkg.Functions[fnName] = fn
		}
	}
}

func removeCommands(ps Packages, api string, ver Version, cmdNames []SpecCommandRef) {
	for _, pc := range ps {
		if pc.Api != api {
			continue
		}
		if pc.Version.Compare(ver) < 0 {
			continue
		}
		for _, cn := range cmdNames {
			fname := TrimGLCmdPrefix(cn.Name)
			if _, ok := pc.Functions[fname]; !ok {
				fmt.Println("Remove cmd: Cmd not found", fname)
			} else {
				delete(pc.Functions, fname)
			}
		}
	}
}

func ParseSpecFile(file string, fs Features) (Packages, error) {
	pacs := make(Packages, 0)

	reg, err := readSpecFile(file)
	if err != nil {
		return nil, err
	}

	functions := commandsToFunctions(reg.Commands)
	tds, err := reg.ParseTypedefs()
	if err != nil {
		return nil, err
	}

	for _, ft := range reg.Features {
		fmt.Println("Feature:", ft.Name, ft.Api, ft.Number)
		version, err := ParseVersion(ft.Number)
		if err != nil {
			return nil, err
		}
		if fs.HasFeature(ft.Api, version) {
			fmt.Println("Adding", ft.Name, ft.Api, ft.Number)
			p := &Package{
				Api:       ft.Api,
				Name:      ft.Api,
				Version:   version,
				TypeDefs:  tds,
				Enums:     make(Enums),
				Functions: make(Functions)}
			pacs = append(pacs, p)
		}
	}

	for _, f := range reg.Features {
		version, err := ParseVersion(f.Number)
		if err != nil {
			return nil, err
		}
		for _, r := range f.Requires {
			addEnums(pacs, f.Api, version, r.Enums, reg.Enums)
		}
		for _, d := range f.Removes {
			removeEnums(pacs, f.Api, version, d.Enums)
		}
		for _, r := range f.Requires {
			addCommands(pacs, f.Api, version, r.Commands, functions)
		}
		for _, d := range f.Removes {
			removeCommands(pacs, f.Api, version, d.Commands)
		}
	}

	return pacs, nil
}

type Feature struct {
	Name     string
	Versions []Version
}

type Features []Feature

func ParseFeatureList(featureStr string) (Features, error) {
	if len(featureStr) == 0 {
		return nil, fmt.Errorf("feature string is empty")
	}
	features := make(Features, 0, 8)
	featureStrs := strings.Split(featureStr, "|")
	for _, f := range featureStrs {
		featver := strings.SplitN(f, ":", 2)
		if len(featver) != 2 {
			return nil, fmt.Errorf("wrong format or version needed: '%s'", featureStr)
		}
		versions := make([]Version, 0, 8)
		versionStrs := strings.Split(featver[1], ",")
		for _, v := range versionStrs {
			version, err := ParseVersion(v)
			if err != nil {
				return nil, err
			}
			versions = append(versions, version)
		}
		features = append(features, Feature{Name: featver[0], Versions: versions})
	}
	return features, nil
}

func (fs Features) HasFeature(name string, ver Version) bool {
	for _, f := range fs {
		if f.Name == name {
			for _, v := range f.Versions {
				if v.Compare(ver) == 0 {
					return true
				}
			}
		}
	}
	return false
}
