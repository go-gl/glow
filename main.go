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

var specURL = "https://cvs.khronos.org/svn/repos/ogl/trunk/doc/registry/public/api"
var specRegexp = regexp.MustCompile(`^(gl|glx|egl|wgl)\.xml$`)

var docURLs = []string{
	"https://cvs.khronos.org/svn/repos/ogl/trunk/ecosystem/public/sdk/docs/man2",
	"https://cvs.khronos.org/svn/repos/ogl/trunk/ecosystem/public/sdk/docs/man3",
	"https://cvs.khronos.org/svn/repos/ogl/trunk/ecosystem/public/sdk/docs/man4"}
var docRegexp = regexp.MustCompile(`^[ew]?gl[^u_].*\.xml$`)

func download(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	xmlDir := flags.String("d", "xml", "XML directory")
	flags.Parse(args)

	specDir := filepath.Join(*xmlDir, "spec")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		log.Fatal("error creating specification output directory:", err)
	}

	docDir := filepath.Join(*xmlDir, "doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		log.Fatal("error creating documentation output directory:", err)
	}

	rev, err := DownloadSvnDir(specURL, specRegexp, specDir)
	if err != nil {
		log.Fatal("error downloading specification files:", err)
	}

	specVersionFile := filepath.Join(specDir, "REVISION")
	if err := ioutil.WriteFile(specVersionFile, []byte(rev), 0644); err != nil {
		log.Fatal("error writing spec revision metadata file:", err)
	}

	for _, url := range docURLs {
		if _, err := DownloadSvnDir(url, docRegexp, docDir); err != nil {
			log.Fatal("error downloading documentation files:", err)
		}
	}
}

func generate(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	xmlDir := flags.String("xml", "xml", "XML directory")
	api := flags.String("api", "", "API to generate (e.g., gl)")
	ver := flags.String("version", "", "API version to generate (e.g., 4.1)")
	profile := flags.String("profile", "", "API profile to generate (e.g., core)")
	addext := flags.String("addext", ".*", "Regular expression of extensions to include (e.g., .*)")
	remext := flags.String("remext", "$^", "Regular expression of extensions to exclude (e.g., .*)")
	restrict := flags.String("restrict", "", "JSON file of symbols to restrict symbol generation")
	lenientInit := flags.Bool("lenientInit", false, "When true missing functions do not fail Init")
	flags.Parse(args)

	version, err := ParseVersion(*ver)
	if err != nil {
		log.Fatal("error parsing version:", err)
	}

	addExtRegexp, err := regexp.Compile(*addext)
	if err != nil {
		log.Fatal("error parsing extension inclusion regexp:", err)
	}

	remExtRegexp, err := regexp.Compile(*remext)
	if err != nil {
		log.Fatal("error parsing extension exclusion regexp:", err)
	}

	packageSpec := &PackageSpec{
		API:          *api,
		Version:      version,
		Profile:      *profile,
		AddExtRegexp: addExtRegexp,
		RemExtRegexp: remExtRegexp,
		LenientInit:  *lenientInit,
	}

	specs, rev := parseSpecifications(*xmlDir)
	docs := parseDocumentation(*xmlDir)

	var pkg *Package
	for _, spec := range specs {
		if spec.HasPackage(packageSpec) {
			pkg = spec.ToPackage(packageSpec)
			pkg.SpecRev = rev
			docs.AddDocs(pkg)
			if len(*restrict) > 0 {
				performRestriction(pkg, *restrict)
			}
			if err := pkg.GeneratePackage(); err != nil {
				log.Fatal("error generating package:", err)
			}
			break
		}
	}
	if pkg == nil {
		log.Fatal("unable to generate package:", packageSpec)
	}
	log.Println("generated package in", pkg.Dir())
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
		log.Fatal("error reading JSON restriction file:", err)
	}
	var r jsonRestriction
	if err = json.Unmarshal(data, &r); err != nil {
		log.Fatal("error parsing JSON restriction file:", err)
	}
	pkg.Filter(lookupMap(r.Enums), lookupMap(r.Functions))
}

func parseSpecifications(xmlDir string) ([]*Specification, string) {
	specDir := filepath.Join(xmlDir, "spec")
	specFiles, err := ioutil.ReadDir(specDir)
	if err != nil {
		log.Fatal("error reading spec file entries:", err)
	}

	specs := make([]*Specification, 0, len(specFiles))
	for _, specFile := range specFiles {
		if !strings.HasSuffix(specFile.Name(), "xml") {
			continue
		}
		spec, err := NewSpecification(filepath.Join(specDir, specFile.Name()))
		if err != nil {
			log.Fatal("error parsing specification:", specFile.Name(), err)
		}
		specs = append(specs, spec)
	}

	rev, err := ioutil.ReadFile(filepath.Join(specDir, "REVISION"))
	if err != nil {
		log.Fatal("error reading spec revision file:", err)
	}

	return specs, string(rev)
}

func parseDocumentation(xmlDir string) Documentation {
	docDir := filepath.Join(xmlDir, "doc")
	docFiles, err := ioutil.ReadDir(docDir)
	if err != nil {
		log.Fatal("error reading doc file entries:", err)
	}

	docs := make([]string, 0, len(docFiles))
	for _, docFile := range docFiles {
		docs = append(docs, filepath.Join(docDir, docFile.Name()))
	}

	doc, err := NewDocumentation(docs)
	if err != nil {
		log.Fatal("error parsing documentation:", err)
	}

	return doc
}

// PackageSpec describes a package to be generated.
type PackageSpec struct {
	API          string
	Version      Version
	Profile      string // If "all" overrides the version spec
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
