package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func generateGoPackages(specsDir string, f []Feature, doc *Documentation) {
	ps, err := ParseSpecFile(filepath.Join(specsDir, openGLSpecFile), f)
	if err != nil {
		fmt.Println("Error while parsing OpenGL specification:", err)
	}
	doc.AnnotatePackages(ps)
	err = ps.GeneratePackages()
	if err != nil {
		fmt.Println("Error while generating OpenGL packages:", err)
	}
}

func downloadSpec(name string, args []string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	src := fs.String("src", "khronos", "Source URL or 'khronos'.")
	odir := fs.String("odir", "glspecs", "Output directory for spec files.")
	fs.Parse(args)
	fmt.Println("Downloading specs ...")
	if *src == "khronos" {
		*src = khronosRegistryBaseURL
	}
	err := downloadAllSpecs(*src, *odir)
	if err != nil {
		fmt.Println("Error while downloading docs:", err)
	}
}

func downloadDoc(name string, args []string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	src := fs.String("src", "khronos", "Source URL or 'khronos'.")
	odir := fs.String("odir", "gldocs", "Output directory for doc files.")
	ver := fs.Int("ver", -1, "Doc version: 2, 3, 4")
	fs.Parse(args)
	if *ver < 2 || *ver > 4 {
		fmt.Println("Invalid doc version:", *ver)
		return
	}
	fmt.Println("Downloading docs ...")
	if *src == "khronos" {
		*src = khronosDocBaseURL
	}
	err := DownloadDocs(*src, fmt.Sprintf("man%d", *ver), *odir)
	if err != nil {
		fmt.Println("Error while downloading docs:", err)
	}
}

func generatePackages(name string, args []string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	sdir := fs.String("sdir", "glspecs", "OpenGL spec directory.")
	ddir := fs.String("ddir", "gldocs", "Documentation directory.")
	feat := fs.String("f", "", "Spec features and version seperated by '|', e.g., -f=gl:2.1|gles1:1.0")
	fs.Parse(args)
	df, err := ParseAllDocs(*ddir)
	if err != nil {
		fmt.Println("Error while parsing docs:", err)
		return
	}
	f, err := ParseFeatureList(*feat)
	if err != nil {
		fmt.Println("Error while parsing feature arguments:", err)
		return
	}
	fmt.Println("Generate Bindings ...")
	generateGoPackages(*sdir, f, df)
}

func printUsage(name string) {
	fmt.Printf("Usage:     %s command [arguments]\n", name)
	fmt.Println("Commands:")
	fmt.Println(" pullspec  Download spec files.")
	fmt.Println(" pulldoc   Download documentation files.")
	fmt.Println(" generate  Generate bindings.")
	fmt.Printf("Type %s <command> -help for a detailed command description.\n", name)
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
	case "pullspec":
		downloadSpec("pullspec", args[1:])
	case "pulldoc":
		downloadDoc("pulldoc", args[1:])
	case "generate":
		generatePackages("generate", args[1:])
	default:
		fmt.Printf("Unknown command: '%s'\n", command)
		printUsage(name)
		os.Exit(-1)
	}
}
