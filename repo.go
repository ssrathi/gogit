package gogit

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type Repo struct {
	GitDir   string
	WorkTree string
}

// Used by 'gogit init' to create a fresh repo.
func NewRepo(path string) (*Repo, error) {
	path, _ = filepath.Abs(path)
	repo := Repo{
		WorkTree: path,
		GitDir:   filepath.Join(path, ".git"),
	}

	// Validate that the WorkTree is either empty or it doesn't exist.
	if IsDirPresent(repo.WorkTree) {
		// Make sure if it empty.
		empty, _ := IsDirEmpty(repo.WorkTree)
		if !empty {
			err := fmt.Errorf("Work-tree %q is not empty", repo.WorkTree)
			return nil, err
		}
	} else {
		// Create the repo work-tree directory.
		if err := os.MkdirAll(repo.WorkTree, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// Create needed subdirectories under the .git directory.
	repo.DirPath(true, "objects")
	repo.DirPath(true, "refs", "tags")
	repo.DirPath(true, "refs", "heads")

	// Create needed files under the .git directory.
	description := []byte("Unnamed repository; edit this file 'description' " +
		"to name the repository.\n")
	desc_file, _ := repo.FilePath(true, "description")
	if err := ioutil.WriteFile(desc_file, description, 0644); err != nil {
		return nil, err
	}

	// HEAD file to point to the master branch initially.
	head_ref := []byte("ref: refs/heads/master\n")
	head_file, _ := repo.FilePath(true, "HEAD")
	if err := ioutil.WriteFile(head_file, head_ref, 0644); err != nil {
		return nil, err
	}

	// Write the default git configuration file. We only support few needed
	// configuration options.
	// NOTE: Go doesn't have a native ini parser. So create it manually.
	default_config := []byte(
		"[core]\n" +
			"\trepositoryformatversion = 0\n" +
			"\tbare = false\n" +
			"\tfilemode = false\n")
	config_file, _ := repo.FilePath(true, "config")
	if err := ioutil.WriteFile(config_file, default_config, 0644); err != nil {
		return nil, err
	}

	// A fresh repo is now cooked. Return it to the caller.
	return &repo, nil
}

// Used by all commands other than "gogit init" to work on an existing repo.
// .git directory can be at given path, or can be at any parent up to rootdir.
func GetRepo(path string) (*Repo, error) {
	for {
		path, _ = filepath.Abs(path)

		// Check if git directory is present.
		GitDir := filepath.Join(path, ".git")
		isPresent := IsDirPresent(GitDir)
		isDir, _ := IsPathDir(GitDir)

		if isPresent && isDir {
			// Found the repo.
			repo := Repo{
				WorkTree: path,
				GitDir:   filepath.Join(path, ".git"),
			}
			return &repo, nil
		}

		// Find the parent directory of the given path.
		parent := filepath.Dir(path)
		if parent == path {
			// This means 'gogit init' was not done before.
			err := errors.New("fatal: not a git repository (or any of the " +
				"parent directories): .git")
			return nil, err
		}

		// Continue on the parent path now.
		path = parent
	}
}

// Get (and optionally create) a directory path inside .git in the repo.
// Example: ["objects", "1e", "ab123"] returns ".git/objects/1e/ab123"
func (r *Repo) DirPath(create bool, paths ...string) (string, error) {
	paths = append([]string{r.GitDir}, paths...)
	path := filepath.Join(paths...)

	// Make sure the path is a directory.
	if IsDirPresent(path) {
		isDir, _ := IsPathDir(path)
		if !isDir {
			err := fmt.Errorf("Path %q is not a directory", path)
			return "", err
		}
	} else {
		// Create the directory if requested.
		if create {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return "", err
			}
		}
	}

	return path, nil
}

// Get a file path inside .git in the repo. Optionally create needed
// directories in the path. Last item in 'paths' is the file name.
func (r *Repo) FilePath(create bool, paths ...string) (string, error) {
	if len(paths) == 0 {
		return "", nil
	}

	// Last element is the filename.
	filename := paths[len(paths)-1]
	paths = paths[:len(paths)-1]
	dirPath, err := r.DirPath(create, paths...)
	if err != nil {
		return "", err
	}

	return filepath.Join(dirPath, filename), nil
}

// Find the data referred by the given sha1 hash and add the data to the
// object as per "Git" specifications.
func (r *Repo) ObjectParse(objHash string) (*GitObject, error) {
	data_file, err := r.FilePath(
		false, "objects", string(objHash[0:2]), string(objHash[2:]))
	if err != nil {
		return nil, err
	}

	// Read the file data and decompress it.
	data, err := ioutil.ReadFile(data_file)
	if err != nil {
		return nil, err
	}

	decompressed := bytes.NewBuffer(data)
	rdr, err := zlib.NewReader(decompressed)
	if err != nil {
		return nil, fmt.Errorf("Malformed object %s: bad data", objHash)
	}
	data, _ = ioutil.ReadAll(rdr)
	rdr.Close()

	// Strip the header from the decompressed data.
	spaceInd := bytes.IndexByte(data, byte(' '))
	objType := string(data[0:spaceInd])
	nullInd := bytes.IndexByte(data, byte('\x00'))

	size, err := strconv.Atoi(string(data[spaceInd+1 : nullInd]))
	if err != nil || len(data)-nullInd-1 != size {
		return nil, fmt.Errorf("Malformed object %s: bad length", objHash)
	}

	// Get the header-stripped data.
	obj := NewObject(objType, data[nullInd+1:])
	return obj, nil
}

// Calculate the sha1 of a git object and optionally write it to a file as
// per "Git" specifications.
func (r *Repo) ObjectWrite(obj *GitObject, write bool) (string, error) {
	// Prepare header
	header := []byte(obj.ObjType + " " + strconv.Itoa(len(obj.ObjData)) + "\x00")
	data := append(header, obj.ObjData...)

	// Compute sha1 of the bytes.
	h := sha1.New()
	h.Write(data)
	sha1hash := hex.EncodeToString(h.Sum(nil))

	if !write {
		return sha1hash, nil
	}

	// Write the compressed data to a path determined by the sha1 hash.
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(data)
	w.Close()

	data_file, err := r.FilePath(
		true, "objects", string(sha1hash[0:2]), string(sha1hash[2:]))
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(data_file, compressed.Bytes(), 0664); err != nil {
		return "", err
	}

	// The data is now written to a file as per Git specification.
	return sha1hash, nil
}
