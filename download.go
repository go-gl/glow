package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	khronosRegistryBaseURL = "https://cvs.khronos.org/svn/repos/ogl/trunk/doc/registry/public/api"
	openGLSpecFile         = "gl.xml"
	eglSpecFile            = "egl.xml"
	wglSpecFile            = "wgl.xml"
	glxSpecFile            = "glx.xml"
)

func DownloadAllSpecs(baseURL, outDir string) error {
	if err := downloadFile(baseURL, openGLSpecFile, outDir, openGLSpecFile); err != nil {
		return err
	}
	if err := downloadFile(baseURL, wglSpecFile, outDir, wglSpecFile); err != nil {
		return err
	}
	if err := downloadFile(baseURL, glxSpecFile, outDir, glxSpecFile); err != nil {
		return err
	}
	if err := downloadFile(baseURL, eglSpecFile, outDir, eglSpecFile); err != nil {
		return err
	}
	return nil
}

func downloadFile(baseURL, fileName, outDir, outFile string) error {
	fullURL := fmt.Sprintf("%s/%s", baseURL, fileName)
	fmt.Printf("Downloading %s...\n", fullURL)
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
