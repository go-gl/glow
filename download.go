package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	khronosDocBaseURL = "https://cvs.khronos.org/svn/repos/ogl/trunk/ecosystem/public/sdk/docs"
	khronosDocDir     = "man"
)

const (
	khronosRegistryBaseURL = "https://cvs.khronos.org/svn/repos/ogl/trunk/doc/registry/public/api"
	openGLSpecFile         = "gl.xml"
	eglSpecFile            = "egl.xml"
	wglSpecFile            = "wgl.xml"
	glxSpecFile            = "glx.xml"
)

func downloadFile(baseURL, fileName, outDir, outFile string) error {
	fullURL := fmt.Sprintf("%s/%s", baseURL, fileName)
	fmt.Printf("Downloading %s ...\n", fullURL)
	r, err := http.Get(fullURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(absPath, 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(absPath, outFile), data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func downloadAllSpecs(baseURL, outDir string) error {
	err := downloadFile(baseURL, openGLSpecFile, outDir, openGLSpecFile)
	if err != nil {
		return err
	}
	err = downloadFile(baseURL, wglSpecFile, outDir, wglSpecFile)
	if err != nil {
		return err
	}
	err = downloadFile(baseURL, glxSpecFile, outDir, glxSpecFile)
	if err != nil {
		return err
	}
	err = downloadFile(baseURL, eglSpecFile, outDir, eglSpecFile)
	if err != nil {
		return err
	}
	return nil
}
