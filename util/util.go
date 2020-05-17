package util

import (
	"io"
	"os"
)

// Check if given directory is empty or not.
func IsDirEmpty(path string) (bool, error) {
	fd, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer fd.Close()

	// Read one item from the path.
	_, err = fd.Readdir(1)
	if err == io.EOF {
		return true, nil
	}

	return false, nil
}

// Check if given directory exists.
func IsDirPresent(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// Check if given path is a directory (not a file).
func IsPathDir(path string) (bool, error) {
	fInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fInfo.IsDir(), nil
}
