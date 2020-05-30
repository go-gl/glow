package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type xmlRegistry struct {
	Types      []xmlType      `xml:"types>type"`
	Enums      []xmlEnumSet   `xml:"enums"`
	Commands   []xmlCommand   `xml:"commands>command"`
	Features   []xmlFeature   `xml:"feature"`
	Extensions []xmlExtension `xml:"extensions>extension"`
}

type xmlType struct {
	Name     string `xml:"name,attr"`
	API      string `xml:"api,attr"`
	Requires string `xml:"requires,attr"`
	Raw      []byte `xml:",innerxml"`
}

type xmlEnumSet struct {
	Enums []xmlEnum `xml:"enum"`
}

type xmlEnum struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
	API   string `xml:"api,attr"`
}

type xmlCommand struct {
	Prototype xmlProto   `xml:"proto"`
	API       string     `xml:"api"`
	Params    []xmlParam `xml:"param"`
}

type xmlSignature []byte

type xmlProto struct {
	Raw xmlSignature `xml:",innerxml"`
}

type xmlParam struct {
	Raw xmlSignature `xml:",innerxml"`
}

type xmlFeature struct {
	API      string       `xml:"api,attr"`
	Number   string       `xml:"number,attr"`
	Requires []xmlRequire `xml:"require"`
	Removes  []xmlRemove  `xml:"remove"`
}

type xmlRequire struct {
	Enums    []xmlEnumRef    `xml:"enum"`
	Commands []xmlCommandRef `xml:"command"`
	Profile  string          `xml:"profile,attr"`
}

type xmlRemove struct {
	Enums    []xmlEnumRef    `xml:"enum"`
	Commands []xmlCommandRef `xml:"command"`
	Profile  string          `xml:"profile,attr"`
}

type xmlEnumRef struct {
	Name string `xml:"name,attr"`
}

type xmlCommandRef struct {
	Name string `xml:"name,attr"`
}

type xmlExtension struct {
	Name      string       `xml:"name,attr"`
	Supported string       `xml:"supported,attr"`
	Requires  []xmlRequire `xml:"require"`
	Removes   []xmlRemove  `xml:"remove"`
}

type specRef struct {
	name string
	api  string
}

type specTypedef struct {
	typedef  *Typedef
	ordinal  int    // Relative declaration order of the typedef
	requires string // Optional name of the typedef required for this typedef
}

type specFunctions map[specRef]*Function
type specEnums map[specRef]*Enum
type specTypedefs map[specRef]*specTypedef

type specAddRemSet struct {
	addedCommands   []string
	addedEnums      []string
	removedCommands []string
	removedEnums    []string
	profile         string
}

// A Specification is a parsed version of an XML registry.
type Specification struct {
	Functions  specFunctions
	Enums      specEnums
	Typedefs   specTypedefs
	Features   []SpecificationFeature
	Extensions []SpecificationExtension
}

// A SpecificationFeature describes a set of commands and enums added and/or
// removed in the context of a particular API and version.
type SpecificationFeature struct {
	API     string
	Version Version
	AddRem  []*specAddRemSet
}

// A SpecificationExtension describes a set of commands and enums added to
// implement an extension.
type SpecificationExtension struct {
	Name      string
	APIRegexp *regexp.Regexp
	AddRem    []*specAddRemSet
}

func readSpecFile(file string) (*xmlRegistry, error) {
	var registry xmlRegistry

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

func parseFunctions(commands []xmlCommand) (specFunctions, error) {
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

		fnRef := specRef{cmdName, cmd.API}
		functions[fnRef] = &Function{
			Name:       cmdName,
			GoName:     TrimAPIPrefix(cmdName),
			Parameters: parameters,
			Return:     cmdReturnType}
	}
	return functions, nil
}

func parseOverloads(functions specFunctions, overloads xmlOverloads) (specFunctions, error) {
	for _, overloadInfo := range overloads.Overloads {
		function := functions.getByName(overloadInfo.Name)
		if function == nil {
			return nil, fmt.Errorf("function <%s> not found to overload", overloadInfo.Name)
		}
		err := overloadFunction(function, overloadInfo)
		if err != nil {
			return nil, err
		}
	}
	return functions, nil
}

func overloadFunction(function *Function, info xmlOverload) error {
	overload := Overload{
		GoName:       function.GoName,
		OverloadName: info.OverloadName,
		Parameters:   make([]Parameter, len(function.Parameters)),
		Return:       function.Return,
	}
	copy(overload.Parameters, function.Parameters)
	for _, change := range info.ParameterChanges {
		if (change.Index < 0) || (change.Index >= len(function.Parameters)) {
			return fmt.Errorf("overload for <%s> has invalid parameter index", info.Name)
		}
		param := &overload.Parameters[change.Index]

		if change.Type != nil {
			_, ctype, err := parseSignature(xmlSignature(change.Type.Signature))
			if err != nil {
				return fmt.Errorf("failed to parse signature of overload for <%s>: %v", info.Name, err)
			}
			// store original type definition as a cast, as this most likely will be needed.
			param.Type.Cast = param.Type.CDefinition
			param.Type.PointerLevel = ctype.PointerLevel
			param.Type.Name = ctype.Name
			param.Type.CDefinition = ctype.CDefinition
		}
		if change.Name != nil {
			param.Name = change.Name.Value
		}
	}
	function.Overloads = append(function.Overloads, overload)
	return nil
}

func parseSignature(signature xmlSignature) (name string, ctype Type, err error) {
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
			}
			// Even if we're not explicitly reading the type we're doing so implicitly
			if !readingName {
				ctype.CDefinition += string(t)
				ctype.PointerLevel += strings.Count(raw, "*")
			}
		case xml.StartElement:
			if t.Name.Local == "ptype" {
				readingType = true
			} else if t.Name.Local == "name" {
				readingName = true
			} else {
				return name, ctype, fmt.Errorf("unexpected signature XML: %s", signature)
			}
		case xml.EndElement:
			if t.Name.Local == "ptype" {
				readingType = false
			} else if t.Name.Local == "name" {
				readingName = false
			}
		}
	}

	// If the XML did not call out the name then parse it out
	if ctype.Name == "" {
		cTypeName := ctype.CDefinition
		cTypeName = strings.Replace(cTypeName, "const", "", -1)
		cTypeName = strings.Replace(cTypeName, "*", "", -1)
		cTypeName = strings.Replace(cTypeName, " ", "", -1)
		ctype.Name = cTypeName
	}

	// Convert statically sized arrays to pointers
	// FIXME: Preserve this type information
	arrayRegexp := regexp.MustCompile("\\[.*]")
	ctype.CDefinition = arrayRegexp.ReplaceAllStringFunc(
		ctype.CDefinition,
		func(_ string) string {
			ctype.PointerLevel += 1
			return "*"
		})

	return name, ctype, nil
}

func parseEnums(enumSets []xmlEnumSet) (specEnums, error) {
	enums := make(specEnums)
	for _, set := range enumSets {
		for _, enum := range set.Enums {
			enumRef := specRef{enum.Name, enum.API}
			enums[enumRef] = &Enum{
				Name:   enum.Name,
				GoName: TrimAPIPrefix(enum.Name),
				Value:  enum.Value}
		}
	}
	return enums, nil
}

func parseTypedefs(types []xmlType) (specTypedefs, error) {
	typedefs := make(specTypedefs)
	for i, xtype := range types {
		typedef, err := parseTypedef(xtype)
		if err != nil {
			return nil, err
		}
		typedefRef := specRef{typedef.Name, xtype.API}
		typedefs[typedefRef] = &specTypedef{
			typedef:  typedef,
			ordinal:  i,
			requires: xtype.Requires}
	}
	return typedefs, nil
}

func parseTypedef(xmlType xmlType) (*Typedef, error) {
	typedef := &Typedef{
		Name:        xmlType.Name,
		CDefinition: ""}

	readingName := false
	decoder := xml.NewDecoder(bytes.NewBuffer(xmlType.Raw))
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
				return typedef, fmt.Errorf("unexpected typedef XML: %s", xmlType.Raw)
			}
		case xml.EndElement:
			if t.Name.Local == "name" {
				readingName = false
			}
		default:
			return typedef, fmt.Errorf("unexpected typedef XML: %s", xmlType.Raw)
		}
	}

	return typedef, nil
}

func parseFeatures(xmlFeatures []xmlFeature) ([]SpecificationFeature, error) {
	features := make([]SpecificationFeature, 0, len(xmlFeatures))
	for _, xmlFeature := range xmlFeatures {
		version, err := ParseVersion(xmlFeature.Number)
		if err != nil {
			return features, err
		}
		feature := SpecificationFeature{
			API:     xmlFeature.API,
			Version: version,
			AddRem:  parseAddRem(xmlFeature.Requires, xmlFeature.Removes),
		}
		features = append(features, feature)
	}
	return features, nil
}

func parseAddRem(requires []xmlRequire, removes []xmlRemove) []*specAddRemSet {
	addRemByProfile := make(map[string]*specAddRemSet)

	addRemForProfile := func(profile string) *specAddRemSet {
		addRem, ok := addRemByProfile[profile]
		if !ok {
			addRem = &specAddRemSet{
				profile:         profile,
				addedEnums:      make([]string, 0),
				addedCommands:   make([]string, 0),
				removedEnums:    make([]string, 0),
				removedCommands: make([]string, 0),
			}
			addRemByProfile[profile] = addRem
		}
		return addRem
	}

	for _, req := range requires {
		addRem := addRemForProfile(req.Profile)
		for _, cmd := range req.Commands {
			addRem.addedCommands = append(addRem.addedCommands, cmd.Name)
		}
		for _, enum := range req.Enums {
			addRem.addedEnums = append(addRem.addedEnums, enum.Name)
		}
	}
	for _, rem := range removes {
		addRem := addRemForProfile(rem.Profile)
		for _, cmd := range rem.Commands {
			addRem.removedCommands = append(addRem.removedCommands, cmd.Name)
		}
		for _, enum := range rem.Enums {
			addRem.removedEnums = append(addRem.removedEnums, enum.Name)
		}
	}

	addRems := make([]*specAddRemSet, 0, len(addRemByProfile))
	for _, addRem := range addRemByProfile {
		addRems = append(addRems, addRem)
	}
	return addRems
}

func parseExtensions(xmlExtensions []xmlExtension) ([]SpecificationExtension, error) {
	extensions := make([]SpecificationExtension, 0, len(xmlExtensions))
	for _, xmlExtension := range xmlExtensions {
		if len(xmlExtension.Removes) > 0 {
			return nil, fmt.Errorf("unexpected extension with removal requirement: %s", xmlExtension)
		}
		extension := SpecificationExtension{
			Name:      xmlExtension.Name,
			APIRegexp: regexp.MustCompile("^(" + xmlExtension.Supported + ")$"),
			AddRem:    parseAddRem(xmlExtension.Requires, xmlExtension.Removes),
		}
		extensions = append(extensions, extension)
	}
	return extensions, nil
}

func (functions specFunctions) get(name, api string) *Function {
	function, ok := functions[specRef{name, api}]
	if ok {
		return function
	}
	return functions[specRef{name, ""}]
}

func (functions specFunctions) getByName(name string) *Function {
	for key, function := range functions {
		if key.name == name {
			return function
		}
	}
	return nil
}

func (enums specEnums) get(name, api string) *Enum {
	enum, ok := enums[specRef{name, api}]
	if ok {
		return enum
	}
	return enums[specRef{name, ""}]
}

func (typedefs specTypedefs) selectRequired(name, api string, requiredTypedefs []*Typedef) {
	specTypedef, ok := typedefs[specRef{name, api}]
	if !ok {
		specTypedef = typedefs[specRef{name, ""}]
	}
	if specTypedef != nil {
		requiredTypedefs[specTypedef.ordinal] = specTypedef.typedef
		if specTypedef.requires != "" {
			typedefs.selectRequired(specTypedef.requires, api, requiredTypedefs)
		}
	}
}

func (feature SpecificationFeature) shouldInclude(pkgSpec *PackageSpec) bool {
	// Ignore mismatched APIs
	if pkgSpec.API != feature.API {
		return false
	}
	// Ignore future versions (unless the version is not relevant)
	if pkgSpec.Version.Compare(feature.Version) < 0 {
		return false
	}
	return true
}

func (extension SpecificationExtension) shouldInclude(pkgSpec *PackageSpec) bool {
	// User extension overrides take precedence
	if pkgSpec.AddExtRegexp != nil && pkgSpec.AddExtRegexp.MatchString(extension.Name) {
		return true
	}
	if pkgSpec.RemExtRegexp != nil && pkgSpec.RemExtRegexp.MatchString(extension.Name) {
		return false
	}
	extensionAPI := pkgSpec.API
	// Special case for GL core profile extension inclusion which uses a pseudo-API
	if pkgSpec.API == "gl" && pkgSpec.Profile == "core" {
		extensionAPI = "glcore"
	}
	return extension.APIRegexp.MatchString(extensionAPI)
}

func (addRem *specAddRemSet) shouldInclude(pkgSpec *PackageSpec) bool {
	// Ignore mismatched profiles (unless the profile is not relevant)
	if addRem.profile != pkgSpec.Profile && addRem.profile != "" {
		return false
	}
	return true
}

// NewSpecification creates a new specification based on an XML registry.
func NewSpecification(registry xmlRegistry, overloads xmlOverloads) (*Specification, error) {
	functions, err := parseFunctions(registry.Commands)
	if err != nil {
		return nil, err
	}

	functions, err = parseOverloads(functions, overloads)
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

	extensions, err := parseExtensions(registry.Extensions)
	if err != nil {
		return nil, err
	}

	spec := &Specification{
		Functions:  functions,
		Enums:      enums,
		Typedefs:   typedefs,
		Features:   features,
		Extensions: extensions,
	}
	return spec, nil
}

// HasPackage determines whether the specification can generate the specified package.
func (spec *Specification) HasPackage(pkgSpec *PackageSpec) bool {
	for _, feature := range spec.Features {
		if pkgSpec.API == feature.API && pkgSpec.Version.Compare(feature.Version) == 0 {
			return true
		}
	}
	return false
}

// ToPackage generates a package from the specification.
func (spec *Specification) ToPackage(pkgSpec *PackageSpec) *Package {
	pkg := &Package{
		API:       pkgSpec.API,
		Name:      pkgSpec.API,
		Version:   pkgSpec.Version,
		Profile:   pkgSpec.Profile,
		TmplDir:   pkgSpec.TmplDir,
		Typedefs:  make([]*Typedef, len(spec.Typedefs)),
		Enums:     make(map[string]*Enum),
		Functions: make(map[string]*PackageFunction),
	}

	// Select the commands and enums relevant to the specified API version
	for _, feature := range spec.Features {
		if !feature.shouldInclude(pkgSpec) {
			continue
		}
		for _, addRem := range feature.AddRem {
			if !addRem.shouldInclude(pkgSpec) {
				continue
			}
			for _, cmd := range addRem.addedCommands {
				pkg.Functions[cmd] = &PackageFunction{
					Function: *spec.Functions.get(cmd, pkg.API),
					Required: !pkgSpec.LenientInit,
				}
			}
			for _, enum := range addRem.addedEnums {
				pkg.Enums[enum] = spec.Enums.get(enum, pkg.API)
			}
			if !pkg.Version.IsAll() {
				for _, cmd := range addRem.removedCommands {
					delete(pkg.Functions, cmd)
				}
				for _, enum := range addRem.removedEnums {
					delete(pkg.Enums, enum)
				}
			}
		}
	}

	// Select the extensions compatible with the specified API version
	for _, extension := range spec.Extensions {
		if !extension.shouldInclude(pkgSpec) {
			continue
		}
		for _, addRem := range extension.AddRem {
			if !addRem.shouldInclude(pkgSpec) {
				continue
			}
			for _, cmd := range addRem.addedCommands {
				_, ok := pkg.Functions[cmd]
				if !ok {
					pkg.Functions[cmd] = &PackageFunction{
						Function: *spec.Functions.get(cmd, pkg.API),
						Required: false,
					}
				}
			}
			for _, enum := range addRem.addedEnums {
				pkg.Enums[enum] = spec.Enums.get(enum, pkg.API)
			}
		}
	}

	// Add the types necessary to declare the functions
	for _, fn := range pkg.Functions {
		spec.Typedefs.selectRequired(fn.Function.Return.Name, pkg.API, pkg.Typedefs)
		for _, param := range fn.Function.Parameters {
			spec.Typedefs.selectRequired(param.Type.Name, pkg.API, pkg.Typedefs)
		}
	}
	typedefCount := 0
	for _, typedef := range pkg.Typedefs {
		if typedef != nil {
			pkg.Typedefs[typedefCount] = typedef
			typedefCount++
		}
	}
	pkg.Typedefs = pkg.Typedefs[:typedefCount]

	return pkg
}
