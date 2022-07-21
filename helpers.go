package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func fileMissing(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
}

func fileMatches(path string, otherContent string) (bool, error) {
	fileMissing, err := fileMissing(path)
	if err != nil {
		return false, err
	}

	// If the file does not exist and the content is empty, the file matches
	// and we should return true.
	if otherContent == "" && fileMissing {
		return true, nil
	}

	// If the file is missing but we have content, the file cannot match
	// so we return false.
	if fileMissing {
		return false, nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	// If the actual content of the file does not match...
	return string(content) == otherContent, nil
}

func getURLContents(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}
