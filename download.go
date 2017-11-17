package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
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

type fileContent struct {
	Type        string   `json:"type"`
	Encoding    string   `json:"encoding"`
	Size        uint     `json:"size"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Content     string   `json:"content"`
	SHA         string   `json:"sha"`
	URL         string   `json:"url"`
	GitURL      string   `json:"git_url"`
	HTMLURL     string   `json:"html_url"`
	DownloadURL string   `json:"download_url"`
	Links       linkUrls `json:"_links"`
}

const maxRequests = 10
const repoOwnerName = "KhronosGroup"

// DownloadGitDir reads an Git repo and downloads all the listed (filtered) files.
func DownloadGitDir(authStr string, repoName string, repoFolder string, filter *regexp.Regexp, outDir string) error {
	client := &http.Client{}
	rootURL := "https://api.github.com/repos/" + repoOwnerName + "/" + repoName + "/contents/" + repoFolder
	req, err := http.NewRequest("GET", rootURL, nil)
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
			url := rootURL + "/" + e.Name
			file := filepath.Join(outDir, e.Name)
			go func(url, file string) {
				defer wg.Done()
				if err := downloadFile(authStr, url, file); err != nil && downloadErr == nil {
					downloadErr = err
				}
				<-c
			}(url, file)
		}
	}
	wg.Wait()

	return downloadErr
}

func downloadFile(authStr, url, filePath string) error {
	log.Println("Downloading", url)

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

	var file fileContent
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(file.Content)

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
