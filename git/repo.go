// Package git implements internal git objects and related APIs.
package git

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ssrathi/gogit/util"
)

// Repo structure to hold the current repository details.
type Repo struct {
	GitDir   string
	WorkTree string
}

// RefEntry keeps a mapping of a reference object with its associated reference.
type RefEntry struct {
	Name    string
	RefHash string
}

// NewRepo is used by 'gogit init' to create a fresh repo.
func NewRepo(path string) (*Repo, error) {
	path, _ = filepath.Abs(path)
	repo := Repo{
		WorkTree: path,
		GitDir:   filepath.Join(path, ".git"),
	}

	// Validate that the WorkTree is either empty or it doesn't exist.
	log.Printf("Creating an empty git repo at path: %q\n", path)
	if util.IsDirPresent(repo.WorkTree) {
		// Make sure if it empty.
		empty, _ := util.IsDirEmpty(repo.WorkTree)
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
	descFile, _ := repo.FilePath(true, "description")
	if err := ioutil.WriteFile(descFile, description, 0644); err != nil {
		return nil, err
	}

	// HEAD file to point to the master branch initially.
	headRef := []byte("ref: refs/heads/master\n")
	headFile, _ := repo.FilePath(true, "HEAD")
	if err := ioutil.WriteFile(headFile, headRef, 0644); err != nil {
		return nil, err
	}

	// Write the default git configuration file. We only support few needed
	// configuration options.
	// NOTE: Go doesn't have a native ini parser. So create it manually.
	defaultConfig := []byte(
		"[core]\n" +
			"\trepositoryformatversion = 0\n" +
			"\tbare = false\n" +
			"\tfilemode = false\n")
	configFile, _ := repo.FilePath(true, "config")
	if err := ioutil.WriteFile(configFile, defaultConfig, 0644); err != nil {
		return nil, err
	}

	// A fresh repo is now cooked. Return it to the caller.
	return &repo, nil
}

// GetRepo is used by all commands other than "gogit init" to work on an existing repo.
// .git directory can be at given path, or can be at any parent up to rootdir.
func GetRepo(path string) (*Repo, error) {
	for {
		path, _ = filepath.Abs(path)

		// Check if git directory is present.
		GitDir := filepath.Join(path, ".git")
		isPresent := util.IsDirPresent(GitDir)
		isDir, _ := util.IsPathDir(GitDir)

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

// DirPath gets (and optionally creates) a directory path inside .git in the repo.
// Example: ["objects", "1e", "ab123"] returns ".git/objects/1e/ab123"
func (r *Repo) DirPath(create bool, paths ...string) (string, error) {
	paths = append([]string{r.GitDir}, paths...)
	path := filepath.Join(paths...)

	// Make sure the path is a directory.
	if util.IsDirPresent(path) {
		isDir, _ := util.IsPathDir(path)
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

// FilePath gets a file path inside .git in the repo. Optionally create needed
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

// ObjectParse finds the data referred by the given sha1 hash and add the data to the
// object as per "Git" specifications.
func (r *Repo) ObjectParse(objHash string) (*GitObject, error) {
	dataFile, err := r.FilePath(
		false, "objects", string(objHash[0:2]), string(objHash[2:]))
	if err != nil {
		return nil, err
	}

	// Read the file data and decompress it.
	data, err := ioutil.ReadFile(dataFile)
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

// ObjectWrite calculates the sha1 of a git object and optionally write it to a
// file as per "Git" specifications.
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

	dataFile, err := r.FilePath(
		true, "objects", string(sha1hash[0:2]), string(sha1hash[2:]))
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(dataFile, compressed.Bytes(), 0664); err != nil {
		return "", err
	}

	// The data is now successfully written to a file as per Git specification.
	return sha1hash, nil
}

// RefResolve resolves a given reference string to an equivalent hash which is
// a valid git object hash.
// Useful to:
//   - Convert a short hash to a list of matching full size hashes.
//   - Convert a symbolic, head or tag reference to a list of matching
//     full size hashes.
//   - Convert a symblic reference to a commit hash (Such as HEAD)
func (r *Repo) RefResolve(ref string) ([]string, error) {
	errmsg := fmt.Sprintf("fatal: ambiguous argument '%s': unknown revision or "+
		"path not in the working tree", ref)

	ref = strings.TrimSpace(ref)
	if ref == "" {
		// Can't do much if nothing is given!
		return nil, fmt.Errorf(errmsg)
	}

	// If HEAD is given, then read the reference inside first.
	if ref == "HEAD" {
		headFile, _ := r.FilePath(false, "HEAD")
		data, err := ioutil.ReadFile(headFile)
		if err != nil {
			return nil, err
		}

		ref = string(data)
		ref = strings.TrimSuffix(ref, "\n")
		ref = ref[5:]
	}

	// If it is a symblic ref, then resolve it by reading the reference files.
	// A symbolic reference is in the format "refs/heads/master".
	for {
		refFile, err := r.FilePath(false, ref)
		if err != nil {
			// Not a symbolic reference if some path of this file is not present
			break
		}

		data, err := ioutil.ReadFile(refFile)
		if err != nil {
			// Not a symbolic reference if its file is not present
			break
		}

		ref = string(data)
		ref = strings.TrimSuffix(ref, "\n")

		if !strings.HasPrefix(ref, "ref: ") {
			// It is not a symblic reference.
			break
		}

		// Resolve the new reference again, till a hash is found.
		ref = ref[5:]
	}

	// Check if the given ref is a valid hexadecimal hash.
	re := regexp.MustCompile(`^[a-fA-F0-9]*$`)
	if !re.MatchString(ref) {
		return nil, fmt.Errorf(errmsg)
	}

	// If the hash is smaller than 4, then return an error. "git" doesn't resolve
	// a hash smaller than 4 characters.
	// Also, git hashes are limited to 40 char (SHA1)
	if len(ref) < 4 || len(ref) > 40 {
		return nil, fmt.Errorf(errmsg)
	}

	// If the hash is given in full size, then return it as is.
	if len(ref) == 40 {
		return []string{ref}, nil
	}

	// There are possibly more than one matches. Collect them in a list by
	// looking at matching files under .git/objects directory.
	objectsPath, err := r.DirPath(true, "objects", ref[0:2])
	if err != nil {
		return nil, err
	}

	// Read all files under this directory and collect all files matching the
	// remaining hash (after first 2 char).
	files, err := ioutil.ReadDir(objectsPath)
	if err != nil {
		return nil, err
	}

	matches := []string{}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ref[2:]) {
			matches = append(matches, ref[0:2]+file.Name())
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf(errmsg)
	}

	return matches, nil
}

// ObjectFind resolves a given reference to a single unambiguous hash.
// The object must match the given type.
func (r *Repo) ObjectFind(ref string) (string, error) {
	matches, err := r.RefResolve(ref)
	if err != nil {
		return "", err
	}

	// If there are more than one matches, then prepare an error with all the
	// matches.
	if len(matches) > 1 {
		msg := fmt.Sprintf("short SHA1 %s is ambiguous\n"+
			"Matching SHA1 list:\n"+
			"%s", ref, strings.Join(matches, "\n"))
		return "", fmt.Errorf(msg)
	}

	return matches[0], nil
}

// ValidateRef strictly validates if a given reference is a valid reference
// in the local repository (or is HEAD). Returns the resolved object hash.
func (r *Repo) ValidateRef(ref string) (string, error) {
	msg := "fatal: '{%s}' - not a valid ref"
	if ref != "HEAD" && !strings.HasPrefix(ref, "refs") {
		return "", fmt.Errorf(msg, ref)
	}

	refHash, err := r.ObjectFind(ref)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf(msg, ref)
	}

	return refHash, nil
}

// GetRefs gets all the references inside the .git directory. This can be
// used by commands such as "gogit show-ref".
func (r *Repo) GetRefs(Pattern string, getHead bool) ([]RefEntry, error) {
	// Read all files inside .git/refs and collect them in a list.
	// If 'Pattern' is given, then filter out all other files.
	// If 'getHead' is given, then get .git/HEAD as well.
	refDir, err := r.DirPath(false, "refs")
	if err != nil {
		return nil, err
	}

	refs := []RefEntry{}
	// Walk function to travese through everything under .git/refs directory.
	var walkFunc = func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			log.Printf("Error in accessing path %s (%v)\n", path, err)
			return walkErr
		}

		if !info.IsDir() {
			if Pattern == "" || Pattern == info.Name() {
				// Get the relative path from the .git directory.
				ref := strings.TrimPrefix(path, (r.GitDir)+"/")
				refHash, err := r.ObjectFind(ref)
				if err != nil {
					return err
				}

				refs = append(refs, RefEntry{ref, refHash})
			}
		}

		return nil
	}

	err = filepath.Walk(refDir, walkFunc)
	if err != nil {
		return nil, err
	}

	// Get HEAD ref if asked for
	if getHead {
		headFile, err := r.FilePath(false, "HEAD")
		if err != nil {
			return nil, err
		}

		// Get the relative path from the .git directory.
		ref := strings.TrimPrefix(headFile, (r.GitDir)+"/")
		refHash, err := r.ObjectFind(ref)
		if err != nil {
			return nil, err
		}

		refs = append(refs, RefEntry{ref, refHash})
	}

	// Sort the entries (git keeps them sorted for display)
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].Name < refs[j].Name
	})
	return refs, nil
}
