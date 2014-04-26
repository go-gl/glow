package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

const (
	specDir = "spec"
	docDir  = "doc"
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
	outDir := flags.String("d", "xml", "Output directory.")
	flags.Parse(args)

	if err := DownloadSvnDir(specURL, specRegexp, filepath.Join(*outDir, specDir)); err != nil {
		log.Fatal("Error downloading specification files:", err)
	}

	for _, url := range docURLs {
		if err := DownloadSvnDir(url, docRegexp, filepath.Join(*outDir, docDir)); err != nil {
			log.Fatal("Error downloading documentation files:", err)
		}
	}
}

func generate(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	specDir := flags.String("d", "xml", "XML directory.")
	featuresSpec := flags.String("f", "", "Spec features and version seperated by '|', e.g., -f=gl:2.1|gles1:1.0")
	flags.Parse(args)

	features, err := ParseFeatureList(*featuresSpec)
	if err != nil {
		log.Fatal("Error parsing feature arguments:", err)
	}

	// Read all spec files up front (parser.go)
	// Find and emit the requested features
	spec, err := ParseSpecFile(filepath.Join(*specDir, "gl.xml"), features)
	if err != nil {
		log.Fatal("Error parsing OpenGL specification:", err)
	}

	err = spec.GeneratePackages()
	if err != nil {
		log.Fatal("Error generating Go packages:", err)
	}
}

func printUsage(name string) {
	fmt.Printf("Usage: %s command [arguments]\n", name)
	fmt.Println("Commands:")
	fmt.Println("  download  Downloads specification and documentation XML files.")
	fmt.Println("  generate  Generates bindings.")
	fmt.Printf("Use %s <command> -help for a detailed command description.\n", name)
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
