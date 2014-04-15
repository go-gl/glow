package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func downloadSpecs(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	specURL := flags.String("url", "khronos", "Source URL or 'khronos'.")
	specDir := flags.String("d", "glspecs", "Output directory for spec files.")
	flags.Parse(args)

	if *specURL == "khronos" {
		*specURL = khronosRegistryBaseURL
	}

	err := DownloadAllSpecs(*specURL, *specDir)
	if err != nil {
		fmt.Println("Error while downloading docs:", err)
	}
}

func generatePackages(name string, args []string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	specDir := flags.String("d", "glspecs", "OpenGL spec file directory.")
	featuresSpec := flags.String("f", "", "Spec features and version seperated by '|', e.g., -f=gl:2.1|gles1:1.0")
	flags.Parse(args)

	features, err := ParseFeatureList(*featuresSpec)
	if err != nil {
		fmt.Println("Error parsing feature arguments:", err)
		return
	}

	spec, err := ParseSpecFile(filepath.Join(*specDir, openGLSpecFile), features)
	if err != nil {
		fmt.Println("Error parsing OpenGL specification:", err)
	}

	err = spec.GeneratePackages()
	if err != nil {
		fmt.Println("Error generating Go packages:", err)
	}
}

func printUsage(name string) {
	fmt.Printf("Usage:     %s command [arguments]\n", name)
	fmt.Println("Commands:")
	fmt.Println(" pullspecs Download specification files.")
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
	case "pullspecs":
		downloadSpecs("pullspecs", args[1:])
	case "generate":
		generatePackages("generate", args[1:])
	default:
		fmt.Printf("Unknown command: '%s'\n", command)
		printUsage(name)
		os.Exit(-1)
	}
}
