// Command glow generates Go OpenGL bindings. See http://github.com/errcw/glow.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func generate(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	dir := importPathToDir("github.com/go-gl/glow")
	var (
		xmlDir      = flags.String("xml", filepath.Join(dir, "xml"), "XML directory")
		tmplDir     = flags.String("tmpl", filepath.Join(dir, "tmpl"), "Template directory")
		outDir      = flags.String("out", "gl", "Output directory")
		api         = flags.String("api", "", "API to generate (e.g., gl)")
		ver         = flags.String("version", "", "API version to generate (e.g., 4.1)")
		profile     = flags.String("profile", "", "API profile to generate (e.g., core)")
		addext      = flags.String("addext", "", "If non-empty, a regular expression describing which extensions to include in addition to those supported by the selected profile; takes precedence over explicit removal")
		remext      = flags.String("remext", "", "If non-empty, a regular expression describing which extensions to exclude")
		restrict    = flags.String("restrict", "", "JSON file of symbols to restrict symbol generation")
		lenientInit = flags.Bool("lenientInit", false, "When true missing functions do not fail Init")
	)
	flags.Parse(args)

	version, err := ParseVersion(*ver)
	if err != nil {
		log.Fatalln("error parsing version:", err)
	}

	var addExtRegexp *regexp.Regexp = nil
	if *addext != "" {
		addExtRegexp, err = regexp.Compile(*addext)
		if err != nil {
			log.Fatalln("error parsing extension inclusion regexp:", err)
		}
	}

	var remExtRegexp *regexp.Regexp = nil
	if *remext != "" {
		remExtRegexp, err = regexp.Compile(*remext)
		if err != nil {
			log.Fatalln("error parsing extension exclusion regexp:", err)
		}
	}

	packageSpec := &PackageSpec{
		API:          *api,
		Version:      version,
		Profile:      *profile,
		TmplDir:      *tmplDir,
		AddExtRegexp: addExtRegexp,
		RemExtRegexp: remExtRegexp,
		LenientInit:  *lenientInit,
	}

	specs := parseSpecifications(*xmlDir)
	docs := parseDocumentation(*xmlDir)

	var pkg *Package
	for _, spec := range specs {
		if spec.HasPackage(packageSpec) {
			pkg = spec.ToPackage(packageSpec)
			docs.AddDocs(pkg)
			if len(*restrict) > 0 {
				performRestriction(pkg, *restrict)
			}
			if err := pkg.GeneratePackage(*outDir); err != nil {
				log.Fatalln("error generating package:", err)
			}
			break
		}
	}
	if pkg == nil {
		log.Fatalln("unable to generate package:", packageSpec)
	}
	log.Println("generated package in", *outDir)
}

// Converts a slice string into a simple lookup map.
func lookupMap(s []string) map[string]bool {
	lookup := make(map[string]bool, len(s))
	for _, str := range s {
		lookup[str] = true
	}
	return lookup
}

type jsonRestriction struct {
	Enums     []string
	Functions []string
}

// Reads the given JSON file path into jsonRestriction and filters the package
// accordingly.
func performRestriction(pkg *Package, jsonPath string) {
	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		log.Fatalln("error reading JSON restriction file:", err)
	}
	var r jsonRestriction
	if err = json.Unmarshal(data, &r); err != nil {
		log.Fatalln("error parsing JSON restriction file:", err)
	}
	pkg.Filter(lookupMap(r.Enums), lookupMap(r.Functions))
}

func parseSpecifications(xmlDir string) []*Specification {
	specDir := filepath.Join(xmlDir, "spec")
	overloadDir := filepath.Join(xmlDir, "overload")
	specFiles, err := ioutil.ReadDir(specDir)
	if err != nil {
		log.Fatalln("error reading spec file entries:", err)
	}

	specs := make([]*Specification, 0, len(specFiles))
	for _, specFile := range specFiles {
		if !strings.HasSuffix(specFile.Name(), "xml") {
			continue
		}

		registry, err := readSpecFile(filepath.Join(specDir, specFile.Name()))
		if err != nil {
			log.Fatalln("error reading XML spec file: ", specFile.Name(), err)
		}
		overloads, err := readOverloadFile(filepath.Join(overloadDir, specFile.Name()))
		if err != nil {
			log.Fatalln("error reading XML overload file: ", specFile.Name(), err)
		}
		spec, err := NewSpecification(*registry, overloads)
		if err != nil {
			log.Fatalln("error parsing specification:", specFile.Name(), err)
		}
		specs = append(specs, spec)
	}

	return specs
}

func parseDocumentation(xmlDir string) Documentation {
	docDir := filepath.Join(xmlDir, "doc")
	docFiles, err := ioutil.ReadDir(docDir)
	if err != nil {
		log.Fatalln("error reading doc file entries:", err)
	}

	docs := make([]string, 0, len(docFiles))
	for _, docFile := range docFiles {
		docs = append(docs, filepath.Join(docDir, docFile.Name()))
	}

	doc, err := NewDocumentation(docs)
	if err != nil {
		log.Fatalln("error parsing documentation:", err)
	}

	return doc
}

// PackageSpec describes a package to be generated.
type PackageSpec struct {
	API          string
	Version      Version
	Profile      string // If "all" overrides the version spec
	TmplDir      string
	AddExtRegexp *regexp.Regexp
	RemExtRegexp *regexp.Regexp
	LenientInit  bool
}

func printUsage(name string) {
	fmt.Printf("Usage: %s command [arguments]\n", name)
	fmt.Println("Commands:")
	fmt.Println("  download  Downloads specification and documentation XML files")
	fmt.Println("  generate  Generates bindings")
	fmt.Printf("Use %s <command> -help for a detailed command description\n", name)
}

func main() {
	name := os.Args[0]
	args := os.Args[1:]

	if len(args) < 1 {
		printUsage(name)
		os.Exit(-1)
	}

	command := args[0]
	switch command {
	case "download":
		download("download", args[1:])
	case "generate":
		generate("generate", args[1:])
	default:
		fmt.Printf("Unknown command: '%s'\n", command)
		printUsage(name)
		os.Exit(-1)
	}
}
