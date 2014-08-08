package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"sync"
)

type svnRoot struct {
	Index svnIndex `xml:"index"`
}

type svnIndex struct {
	Revision string     `xml:"rev,attr"`
	Entries  []svnEntry `xml:"file"`
}

type svnEntry struct {
	Name string `xml:"href,attr"`
}

const maxRequests = 10

// DownloadSvnDir reads an SVN XML directory index and downloads all the listed (filtered) files. It
// returns the SVN revision at which the files were downloaded.
func DownloadSvnDir(svnDirURL string, filter *regexp.Regexp, outDir string) (string, error) {
	response, err := http.Get(svnDirURL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var root svnRoot
	if err := xml.NewDecoder(response.Body).Decode(&root); err != nil {
		return "", err
	}
	index := root.Index

	var downloadErr error

	wg := new(sync.WaitGroup)
	c := make(chan int, maxRequests)
	for _, e := range index.Entries {
		if filter.MatchString(e.Name) {
			c <- 1
			wg.Add(1)
			url := svnDirURL + "/" + e.Name
			file := filepath.Join(outDir, e.Name)
			go func(url, file string) {
				defer wg.Done()
				if err := downloadFile(url, file); err != nil && downloadErr == nil {
					downloadErr = err
				}
				<-c
			}(url, file)
		}
	}
	wg.Wait()

	return index.Revision, downloadErr
}

func downloadFile(url, file string) error {
	log.Println("Downloading", url)

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, body, 0644)
	if err != nil {
		return err
	}

	return nil
}
