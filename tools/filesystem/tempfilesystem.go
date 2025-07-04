package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const httpGetTimeout = 20 * time.Second

var errBadHTTPStatus = errors.New("bad http status")

type TempFileSystem struct {
	files []*os.File
}

func NewTempFileSystem() *TempFileSystem {
	return &TempFileSystem{files: []*os.File{}}
}

func (fs *TempFileSystem) DownloadTempFile(dir, pattern string, url *url.URL) (string, error) {
	bytes, err := getBytesFromURL(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file from %s: %w", url, err)
	}

	tmpFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file with pattern %s: %w", pattern, err)
	}
	defer tmpFile.Close()
	fs.files = append(fs.files, tmpFile)
	if _, err := tmpFile.Write(bytes); err != nil {
		return "", fmt.Errorf("failed to write to temp file %s: %w", tmpFile.Name(), err)
	}
	return tmpFile.Name(), nil
}

// FileExists checks if a file exists at the given filePath and is a regular file.
func (fs *TempFileSystem) FileExists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File does not exist
		}
		return false, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}
	return info.Mode().IsRegular(), nil
}

func (fs *TempFileSystem) RemoveTempFiles() []error {
	var errs []error
	for _, file := range fs.files {
		err := os.Remove(file.Name())
		if err != nil {
			errs = append(errs, err)
		}
	}
	fs.files = []*os.File{}
	return errs
}

func getBytesFromURL(url *url.URL) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpGetTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP GET request for %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET request failed for %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d for GET request to %s: %w", resp.StatusCode, url, errBadHTTPStatus)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return data, nil
}
