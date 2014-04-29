package main

import (
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
		log.Fatal("Error creating specification output directory:", err)
	}

	docDir := filepath.Join(*xmlDir, "doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		log.Fatal("Error creating documentation output directory:", err)
	}

	if err := DownloadSvnDir(specURL, specRegexp, specDir); err != nil {
		log.Fatal("Error downloading specification files:", err)
	}

	for _, url := range docURLs {
		if err := DownloadSvnDir(url, docRegexp, docDir); err != nil {
			log.Fatal("Error downloading documentation files:", err)
		}
	}
}

func generate(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	xmlDir := flags.String("d", "xml", "XML directory")
	pkgs := flags.String("g", "", "Packages to generate in a list of 'api@version' (e.g., gl@4.4) or 'all'")
	flags.Parse(args)

	specDir := filepath.Join(*xmlDir, "spec")
	specFiles, err := ioutil.ReadDir(specDir)
	if err != nil {
		log.Fatal("Error reading spec file entries:", err)
	}

	specs := make([]*Specification, 0, len(specFiles))
	for _, specFile := range specFiles {
		spec, err := NewSpecification(filepath.Join(specDir, specFile.Name()))
		if err != nil {
			log.Fatal("Error parsing specification:", specFile.Name(), err)
		}
		specs = append(specs, spec)
	}

	packageSpecs, err := parsePackageSpecs(*pkgs, specs)
	if err != nil {
		log.Fatal("Error parsing generation arguments:", err)
	}

	for _, pkgSpec := range packageSpecs {
		generated := false
		for _, spec := range specs {
			if spec.HasPackage(pkgSpec) {
				log.Println("Generating package", pkgSpec.Api, pkgSpec.Version)
				if err := spec.ToPackage(pkgSpec).GeneratePackage(); err != nil {
					log.Fatal("Error generating package:", err)
				}
				generated = true
				break
			}
		}
		if !generated {
			log.Fatal("Unable to generate package:", pkgSpec)
		}
	}
}

type PackageSpec struct {
	Api     string
	Version Version
}

func (pkgSpec PackageSpec) String() string {
	return fmt.Sprintf("%s %s", pkgSpec.Api, pkgSpec.Version)
}

func parsePackageSpecs(specStrs string, specs []*Specification) ([]PackageSpec, error) {
	pkgSpecs := make([]PackageSpec, 0)
	if specStrs == "all" {
		for _, spec := range specs {
			for _, feature := range spec.Features {
				pkgSpecs = append(pkgSpecs, PackageSpec{feature.Api, feature.Version})
			}
		}
	} else {
		for _, specStr := range strings.Split(specStrs, ",") {
			apiVersion := strings.Split(specStr, "@")
			if len(apiVersion) != 2 {
				return nil, fmt.Errorf("Error parsing generation specification:", specStr)
			}
			api := apiVersion[0]
			version, err := ParseVersion(apiVersion[1])
			if err != nil {
				return nil, err
			}
			pkgSpecs = append(pkgSpecs, PackageSpec{api, version})
		}
	}
	return pkgSpecs, nil
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
