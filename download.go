package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

// DownloadFile downloads a single file from a URL to a specified location.
func DownloadFile(url, file string) error {
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
