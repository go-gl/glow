// Command glow generates Go OpenGL bindings. See http://github.com/errcw/glow.
package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var specRepoName = "OpenGL-Registry"
var specRepoFolder = "xml"
var specRegexp = regexp.MustCompile(`^(gl|glx|wgl)\.xml$`)
var eglRepoName = "EGL-Registry"
var eglRepoFolder = "api"
var eglRegexp = regexp.MustCompile(`^(egl)\.xml$`)
var docRepoName = "OpenGL-Refpages"
var docRepoFolders = []string{
	"es1.1",
	"es2.0",
	"es3.0",
	"es3.1",
	"es3",
	"gl2.1",
	"gl4",
}
var docRegexp = regexp.MustCompile(`^[ew]?gl[^u_].*\.xml$`)

func auth(username string, password string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)

	if err != nil {
		return "", err
	}

	autStr := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))
	req.Header.Add("Authorization", autStr)
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("GitHub authorization failed")
	}

	defer resp.Body.Close()
	return autStr, nil
}

func download(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	xmlDir := flags.String("d", "xml", "XML directory")
	flags.Parse(args)

	specDir := filepath.Join(*xmlDir, "spec")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		log.Fatalln("error creating specification output directory:", err)
	}

	docDir := filepath.Join(*xmlDir, "doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		log.Fatalln("error creating documentation output directory:", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter GitHub username: ")
	input, _ := reader.ReadString('\n')
	username := strings.Replace(input, "\n", "", -1)
	fmt.Print("Enter GitHub password: ")
	input, _ = reader.ReadString('\n')
	password := strings.Replace(input, "\n", "", -1)

	authStr, err := auth(username, password)

	if err != nil {
		panic(err)
	}

	err = DownloadGitDir(authStr, specRepoName, specRepoFolder, specRegexp, specDir)
	if err != nil {
		log.Fatalln("error downloading specification files:", err)
	}

	err = DownloadGitDir(authStr, eglRepoName, eglRepoFolder, eglRegexp, specDir)
	if err != nil {
		log.Fatalln("error downloading egl file:", err)
	}

	for _, folder := range docRepoFolders {
		if err := DownloadGitDir(authStr, docRepoName, folder, docRegexp, docDir); err != nil {
			log.Fatalln("error downloading documentation files:", err)
		}
	}
}

func generate(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	var (
		xmlDir      = flags.String("xml", importPathToDir("github.com/go-gl/glow/xml"), "XML directory")
		tmplDir     = flags.String("tmpl", importPathToDir("github.com/go-gl/glow/tmpl"), "Template directory")
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

func parseSpecifications(xmlDir string) ([]*Specification, string) {
	specDir := filepath.Join(xmlDir, "spec")
	specFiles, err := ioutil.ReadDir(specDir)
	if err != nil {
		log.Fatalln("error reading spec file entries:", err)
	}

	specs := make([]*Specification, 0, len(specFiles))
	for _, specFile := range specFiles {
		if !strings.HasSuffix(specFile.Name(), "xml") {
			continue
		}
		spec, err := NewSpecification(filepath.Join(specDir, specFile.Name()))
		if err != nil {
			log.Fatalln("error parsing specification:", specFile.Name(), err)
		}
		specs = append(specs, spec)
	}

	rev, err := ioutil.ReadFile(filepath.Join(specDir, "REVISION"))
	if err != nil {
		log.Fatalln("error reading spec revision file:", err)
	}

	return specs, string(rev)
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
