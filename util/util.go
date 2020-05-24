// Package util implements miscellaneous utility APIs.
package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Check is a helper function to exit on irrecoverable error.
func Check(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("%s:%d %s", filepath.Base(file), line, err)
		fmt.Println(err)
		os.Exit(1)
	}
}

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

// IsPathPresent checks if given path (dir or file) exists.
func IsPathPresent(path string) bool {
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
