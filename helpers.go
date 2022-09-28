package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	errUnexpectedStatusCode = fmt.Errorf("unexpected status code")
)

func fileMissing(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}

	if os.IsNotExist(err) {
		return true, nil
	}

	return false, fmt.Errorf("error finding file %s: %w", path, err)
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

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return false, fmt.Errorf("error reading file %s: %w", path, err)
	}

	// If the actual content of the file does not match...
	return string(content) == otherContent, nil
}

func getURLContents(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// The DefaultClient is used here but should be replaced with a custom client.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get URL contents: %w", err)
	}

	defer func() {
		if bodyCloseErr := resp.Body.Close(); bodyCloseErr != nil {
			log.Println("error closing response body:", bodyCloseErr)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d", errUnexpectedStatusCode, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body of %s: %w", url, err)
	}

	return string(body), err
}
