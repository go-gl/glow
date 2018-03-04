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
	"sync"
)

type linkUrls struct {
	Self string `json:"self"`
	Git  string `json:"git"`
	HTML string `json:"html"`
}

type dirContent struct {
	Type        string   `json:"type"`
	Size        uint     `json:"size"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	SHA         string   `json:"sha"`
	URL         string   `json:"url"`
	GitURL      string   `json:"git_url"`
	HTMLURL     string   `json:"html_url"`
	DownloadURL string   `json:"download_url"`
	Links       linkUrls `json:"_links"`
}

type blobContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	URL      string `json:"url"`
	SHA      string `json:"sha"`
	Size     uint   `json:"size"`
}

const maxRequests = 10
const repoOwnerName = "KhronosGroup"

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

func validatedAuthHeader(username string, password string) (string, error) {
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

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("GitHub authorization failed")
	}

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
	username := strings.Trim(input, "\n")
	fmt.Print("Enter GitHub password: ")
	input, _ = reader.ReadString('\n')
	password := strings.Trim(input, "\n")

	authHeader, err := validatedAuthHeader(username, password)
	if err != nil {
		log.Fatalln("error with user authorization:", err)
	}

	err = DownloadGitDir(authHeader, specRepoName, specRepoFolder, specRegexp, specDir)
	if err != nil {
		log.Fatalln("error downloading specification files:", err)
	}

	err = DownloadGitDir(authHeader, eglRepoName, eglRepoFolder, eglRegexp, specDir)
	if err != nil {
		log.Fatalln("error downloading egl file:", err)
	}

	for _, folder := range docRepoFolders {
		if err := DownloadGitDir(authHeader, docRepoName, folder, docRegexp, docDir); err != nil {
			log.Fatalln("error downloading documentation files:", err)
		}
	}
}

// DownloadGitDir reads an Git repo and downloads all the listed (filtered) files.
func DownloadGitDir(authStr string, repoName string, repoFolder string, filter *regexp.Regexp, outDir string) error {
	client := &http.Client{}
	rootDirURL := "https://api.github.com/repos/" + repoOwnerName + "/" + repoName + "/contents/" + repoFolder
	rootBlobURL := "https://api.github.com/repos/" + repoOwnerName + "/" + repoName + "/git/blobs/"
	req, err := http.NewRequest("GET", rootDirURL, nil)
	req.Header.Add("Authorization", authStr)
	req.Header.Add("User-Agent", "go-gl/glow")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var repoContent []dirContent
	if err := json.NewDecoder(resp.Body).Decode(&repoContent); err != nil {
		return err
	}

	var downloadErr error

	wg := new(sync.WaitGroup)
	c := make(chan int, maxRequests)
	for _, e := range repoContent {
		if filter.MatchString(e.Name) {
			c <- 1
			wg.Add(1)
			file := filepath.Join(outDir, e.Name)
			url := rootBlobURL + e.SHA
			go func(url, file string) {
				defer wg.Done()
				if err := downloadBlob(authStr, url, file); err != nil && downloadErr == nil {
					downloadErr = err
				}
				<-c
			}(url, file)
		}
	}
	wg.Wait()

	return downloadErr
}

func downloadBlob(authStr, url, filePath string) error {
	log.Println("Downloading", filePath)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", authStr)
	req.Header.Add("User-Agent", "go-gl/glow")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var blob blobContent
	if err := json.NewDecoder(resp.Body).Decode(&blob); err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(blob.Content)

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
