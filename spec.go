package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type specRegistry struct {
	Types      []specType      `xml:"types>type"`
	Enums      []specEnumSet   `xml:"enums"`
	Commands   []specCommand   `xml:"commands>command"`
	Features   []specFeature   `xml:"feature"`
	Extensions []specExtension `xml:"extensions>extension"`
}

type specType struct {
	Name     string `xml:"name,attr"`
	Api      string `xml:"api,attr"`
	Requires string `xml:"requires,attr"`
	Raw      []byte `xml:",innerxml"`
}

type specEnumSet struct {
	Enums []specEnum `xml:"enum"`
}

type specEnum struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
	Api   string `xml:"api,attr"`
}

type specCommand struct {
	Prototype specProto   `xml:"proto"`
	Api       string      `xml:"api"`
	Params    []specParam `xml:"param"`
}

type specSignature []byte

type specProto struct {
	Raw specSignature `xml:",innerxml"`
}

type specParam struct {
	Raw specSignature `xml:",innerxml"`
}

type specFeature struct {
	Api      string        `xml:"api,attr"`
	Number   string        `xml:"number,attr"`
	Requires []specRequire `xml:"require"`
	Removes  []specRemove  `xml:"remove"`
}

type specRequire struct {
	Enums    []specEnumRef    `xml:"enum"`
	Commands []specCommandRef `xml:"command"`
}

type specRemove struct {
	Enums    []specEnumRef    `xml:"enum"`
	Commands []specCommandRef `xml:"command"`
}

type specEnumRef struct {
	Name string `xml:"name,attr"`
}

type specCommandRef struct {
	Name string `xml:"name,attr"`
}

type specExtension struct {
	Name      string        `xml:"name,attr"`
	Supported string        `xml:"supported,attr"`
	Requires  []specRequire `xml:"require"`
}

type specRef struct {
	Name string
	Api  string
}

type specTypedef struct {
	ref      specRef
	typedef  *Typedef
	required bool
}

type specFunctions map[specRef]*Function
type specEnums map[specRef]*Enum

// Use an ordered list to preserve the XML declaration ordering which encodes implicit dependencies
type specTypedefs []specTypedef

// Parsed version of the XML specification
type Specification struct {
	Functions specFunctions
	Enums     specEnums
	Typedefs  specTypedefs
	Features  []SpecificationFeature
}

type SpecificationFeature struct {
	Api     string
	Version Version

	AddedEnums      []string
	AddedCommands   []string
	RemovedEnums    []string
	RemovedCommands []string
}

func readSpecFile(file string) (*specRegistry, error) {
	var registry specRegistry

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = xml.NewDecoder(f).Decode(&registry)
	if err != nil {
		return nil, err
	}

	return &registry, nil
}

func parseFunctions(commands []specCommand) (specFunctions, error) {
	functions := make(specFunctions)
	for _, cmd := range commands {
		cmdName, cmdReturnType, err := parseSignature(cmd.Prototype.Raw)
		if err != nil {
			return functions, err
		}

		parameters := make([]Parameter, 0, len(cmd.Params))
		for _, param := range cmd.Params {
			paramName, paramType, err := parseSignature(param.Raw)
			if err != nil {
				return functions, err
			}
			parameter := Parameter{
				Name: paramName,
				Type: paramType}
			parameters = append(parameters, parameter)
		}

		fnRef := specRef{cmdName, cmd.Api}
		functions[fnRef] = &Function{
			Name:       cmdName,
			GoName:     TrimGLCmdPrefix(cmdName),
			Parameters: parameters,
			Return:     cmdReturnType}
	}
	return functions, nil
}

func parseSignature(signature specSignature) (string, Type, error) {
	name := ""
	ctype := Type{}

	readingName := false
	readingType := false

	decoder := xml.NewDecoder(bytes.NewBuffer(signature))
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
			raw := strings.Trim(string(t), " ")
			if readingName {
				name = raw
			} else if readingType {
				ctype.Name = raw
			} else {
				if strings.Contains(raw, "void") {
					ctype.Name = "void"
				}
				ctype.PointerLevel += strings.Count(raw, "*")
			}
			if !readingName {
				ctype.CDefinition += string(t)
			}
		case xml.StartElement:
			if t.Name.Local == "ptype" {
				readingType = true
			} else if t.Name.Local == "name" {
				readingName = true
			} else {
				return name, ctype, fmt.Errorf("Unexpected signature XML: %s", signature)
			}
		case xml.EndElement:
			if t.Name.Local == "ptype" {
				readingType = false
			} else if t.Name.Local == "name" {
				readingName = false
			}
		}
	}
	return name, ctype, nil
}

func parseEnums(enumSets []specEnumSet) (specEnums, error) {
	enums := make(specEnums)
	for _, set := range enumSets {
		for _, enum := range set.Enums {
			enumRef := specRef{enum.Name, enum.Api}
			enums[enumRef] = &Enum{
				Name:   enum.Name,
				GoName: TrimGLEnumPrefix(enum.Name),
				Value:  enum.Value}
		}
	}
	return enums, nil
}

func parseTypedefs(specTypes []specType) (specTypedefs, error) {
	typedefs := make(specTypedefs, 0, len(specTypes))
	for _, specType := range specTypes {
		typedef, err := parseTypedef(specType)
		if err != nil {
			return nil, err
		}
		specTd := specTypedef{
			specRef{typedef.Name, specType.Api},
			typedef,
			false}
		typedefs = append(typedefs, specTd)
	}
	return typedefs, nil
}

func parseTypedef(specType specType) (*Typedef, error) {
	typedef := &Typedef{
		Name:        specType.Name,
		Api:         specType.Api,
		Requires:    specType.Requires,
		CDefinition: ""}

	readingName := false
	decoder := xml.NewDecoder(bytes.NewBuffer(specType.Raw))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return typedef, err
		}
		switch t := token.(type) {
		case xml.CharData:
			raw := string(t)
			typedef.CDefinition += raw
			if readingName {
				typedef.Name = raw
			}
		case xml.StartElement:
			if t.Name.Local == "name" {
				readingName = true
			} else if t.Name.Local == "apientry" {
				typedef.CDefinition += "APIENTRY"
			} else {
				return typedef, fmt.Errorf("Unexpected typedef XML: %s", specType.Raw)
			}
		case xml.EndElement:
			if t.Name.Local == "name" {
				readingName = false
			}
		default:
			return typedef, fmt.Errorf("Unexpected typedef XML: %s", specType.Raw)
		}
	}

	return typedef, nil
}

func parseFeatures(features []specFeature) ([]SpecificationFeature, error) {
	specFeatures := make([]SpecificationFeature, 0, len(features))
	for _, feature := range features {
		version, err := ParseVersion(feature.Number)
		if err != nil {
			return specFeatures, err
		}

		specFeature := SpecificationFeature{
			Api:             feature.Api,
			Version:         version,
			AddedEnums:      make([]string, 0),
			AddedCommands:   make([]string, 0),
			RemovedEnums:    make([]string, 0),
			RemovedCommands: make([]string, 0),
		}

		for _, req := range feature.Requires {
			for _, cmd := range req.Commands {
				specFeature.AddedCommands = append(specFeature.AddedCommands, cmd.Name)
			}
			for _, enum := range req.Enums {
				specFeature.AddedEnums = append(specFeature.AddedEnums, enum.Name)
			}
		}

		for _, rem := range feature.Removes {
			for _, cmd := range rem.Commands {
				specFeature.RemovedCommands = append(specFeature.RemovedCommands, cmd.Name)
			}
			for _, enum := range rem.Enums {
				specFeature.RemovedEnums = append(specFeature.RemovedEnums, enum.Name)
			}
		}

		specFeatures = append(specFeatures, specFeature)
	}
	return specFeatures, nil
}

func (functions specFunctions) get(name, api string) *Function {
	function, ok := functions[specRef{name, api}]
	if ok {
		return function
	}
	return functions[specRef{name, ""}]
}

func (enums specEnums) get(name, api string) *Enum {
	enum, ok := enums[specRef{name, api}]
	if ok {
		return enum
	}
	return enums[specRef{name, ""}]
}

func (typedefs specTypedefs) markRequired(name, api string) {
	for _, considerApi := range []bool{true, false} {
		for i, typedef := range typedefs {
			if typedef.ref.Name == name && (typedef.ref.Api == api || !considerApi) {
				typedefs[i].required = true
				if typedef.typedef.Requires != "" {
					typedefs.markRequired(typedef.typedef.Requires, api)
				}
				return
			}
		}
	}
}

// NewSpecification creates a new specification based on an XML file.
func NewSpecification(file string) (*Specification, error) {
	registry, err := readSpecFile(file)
	if err != nil {
		return nil, err
	}

	functions, err := parseFunctions(registry.Commands)
	if err != nil {
		return nil, err
	}

	enums, err := parseEnums(registry.Enums)
	if err != nil {
		return nil, err
	}

	typedefs, err := parseTypedefs(registry.Types)
	if err != nil {
		return nil, err
	}

	features, err := parseFeatures(registry.Features)
	if err != nil {
		return nil, err
	}

	spec := &Specification{
		Functions: functions,
		Enums:     enums,
		Typedefs:  typedefs,
		Features:  features,
	}
	return spec, nil
}

// HasPackage determines whether the specification can generate the specified package.
func (spec *Specification) HasPackage(pkgSpec PackageSpec) bool {
	for _, feature := range spec.Features {
		if pkgSpec.Api == feature.Api && pkgSpec.Version.Compare(feature.Version) == 0 {
			return true
		}
	}
	return false
}

// ToPackage generates a package from the specification.
func (spec *Specification) ToPackage(pkgSpec PackageSpec) *Package {
	pkg := &Package{
		Api:       pkgSpec.Api,
		Name:      pkgSpec.Api,
		Version:   pkgSpec.Version,
		Typedefs:  make([]*Typedef, 0),
		Enums:     make(Enums),
		Functions: make(Functions)}

	// Select the commands and enums relevant to the specified API version
	for _, feature := range spec.Features {
		// Skip features from a different API
		if pkg.Api != feature.Api {
			continue
		}
		// Skip features from a later version than the package
		if pkg.Version.Compare(feature.Version) < 0 {
			continue
		}

		for _, enum := range feature.AddedEnums {
			pkg.Enums[enum] = spec.Enums.get(enum, pkg.Api)
		}
		for _, cmd := range feature.AddedCommands {
			pkg.Functions[cmd] = spec.Functions.get(cmd, pkg.Api)
		}

		for _, enum := range feature.RemovedEnums {
			delete(pkg.Enums, enum)
		}
		for _, cmd := range feature.RemovedCommands {
			delete(pkg.Functions, cmd)
		}
	}

	// Add the types necessary to declare the functions
	specTypedefs := make(specTypedefs, len(spec.Typedefs))
	copy(specTypedefs, spec.Typedefs)
	for _, fn := range pkg.Functions {
		specTypedefs.markRequired(fn.Return.Name, pkg.Api)
		for _, param := range fn.Parameters {
			specTypedefs.markRequired(param.Type.Name, pkg.Api)
		}
	}
	for _, specTypedef := range specTypedefs {
		if specTypedef.required {
			pkg.Typedefs = append(pkg.Typedefs, specTypedef.typedef)
		}
	}

	return pkg
}
