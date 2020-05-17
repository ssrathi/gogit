package util

import (
	"io"
	"os"
)

// IsDirEmpty checks if given directory is empty or not.
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

// IsDirPresent checks if given directory exists.
func IsDirPresent(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// IsPathDir checks if given path is a directory (not a file).
func IsPathDir(path string) (bool, error) {
	fInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fInfo.IsDir(), nil
}
