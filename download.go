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

type svnIndex struct {
	XMLName xml.Name   `xml:"svn"`
	Entries []svnEntry `xml:"index>file"`
}

type svnEntry struct {
	Name string `xml:"href,attr"`
}

// DownloadSvnDir reads an SVN XML directory index and downloads all the listed (filtered) files
func DownloadSvnDir(svnDirUrl string, filter *regexp.Regexp, outDir string) error {
	response, err := http.Get(svnDirUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var index svnIndex
	if err := xml.NewDecoder(response.Body).Decode(&index); err != nil {
		return err
	}

	var downloadErr error = nil

	wg := new(sync.WaitGroup)
	for _, e := range index.Entries {
		if filter.MatchString(e.Name) {
			url := svnDirUrl + "/" + e.Name
			file := filepath.Join(outDir, e.Name)
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := downloadFile(url, file); err != nil {
					// Choose an arbitrary error to report
					downloadErr = err
				}
			}()
		}
	}
	wg.Wait()

	return downloadErr
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
