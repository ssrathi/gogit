package gogit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Repo struct {
	GitDir   string
	WorkTree string
}

// Used by 'gogit init' to create a fresh repo.
func NewRepo(path string) (*Repo, error) {
	path, _ = filepath.Abs(path)

	repo := new(Repo)
	repo.WorkTree = filepath.Clean(path)
	repo.GitDir = filepath.Join(repo.WorkTree, ".git")

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
	return repo, nil
}

// Used by all commands other than "gogit init" to work on an existing repo.
// .git directory can be at given path, or can be at any parent up to rootdir.
func GetRepo(path string) (*Repo, error) {
	for {
		path, _ = filepath.Abs(path)
		fmt.Printf("path: %s\n", path)

		// Check if git directory is present.
		GitDir := filepath.Join(path, ".git")
		isPresent := IsDirPresent(GitDir)
		isDir, _ := IsPathDir(GitDir)

		if isPresent && isDir {
			// Found the repo.
			repo := new(Repo)
			repo.WorkTree = path
			repo.GitDir = filepath.Join(repo.WorkTree, ".git")
			return repo, nil
		}

		// Find the parent directory of the given path.
		parent := filepath.Dir(path)
		if parent == path {
			// This means 'gogit init' was not done before.
			err := fmt.Errorf(".git not found anywhere in path %q. "+
				"Use 'init' command to create a repository first.", path)
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
