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
func (r *Repo) ObjectParse(objHash string) (*Object, error) {
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
func (r *Repo) ObjectWrite(obj *Object, write bool) (string, error) {
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

// RefResolve converts a symbolic reference to its object hash.
func (r *Repo) RefResolve(path string) (string, error) {
	for {
		refFile, err := r.FilePath(false, path)
		if err != nil {
			// Not a symbolic reference if some path of this file is not present
			return "", err
		}

		data, err := ioutil.ReadFile(refFile)
		if err != nil {
			// Not a symbolic reference if its file is not present
			return "", err
		}

		ref := string(data)
		ref = strings.TrimSuffix(ref, "\n")

		if !strings.HasPrefix(ref, "ref: ") {
			// It is not a symblic reference.
			return ref, nil
		}

		// Resolve the new reference again, till a hash is found.
		path = ref[5:]
	}
}

// GetRefs gets all the references inside the .git directory. This can be
// used by commands such as "gogit show-ref".
func (r *Repo) GetRefs(pattern string, getHead bool) ([]RefEntry, error) {
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
			// Get the relative path from the .git directory.
			ref := strings.TrimPrefix(path, (r.GitDir)+"/")
			log.Printf("Working on ref: %s\n", ref)

			if pattern != "" {
				if !strings.HasSuffix(ref, pattern) {
					// Given pattern is not applicable to this reference.
					log.Printf("ref %s doesn't end on pattern %s", ref, pattern)
					return nil
				}

				// Find the starting point of the pattern.
				li := strings.LastIndex(ref, pattern)
				if li != 0 && ref[li-1] != byte('/') {
					// Given pattern doesn't match this reference.
					log.Printf("ref %s doesn't have a separator at index %d\n", ref, li-1)
					return nil
				}
			}

			// This is a valid reference. It either matched the pattern or
			// a pattern is not provided.
			// Get the relative path from the .git directory.
			log.Printf("Found %s as a valid reference\n", ref)
			refHash, err := r.RefResolve(ref)
			if err != nil {
				return err
			}

			refs = append(refs, RefEntry{ref, refHash})
		}

		return nil
	}

	err = filepath.Walk(refDir, walkFunc)
	if err != nil {
		return nil, err
	}

	// Get HEAD ref if asked for
	if getHead {
		headHash, err := r.RefResolve("HEAD")
		if err != nil {
			return nil, err
		}

		log.Println("Found valid HEAD reference for HEAD")
		refs = append(refs, RefEntry{"HEAD", headHash})
	}

	// Sort the entries (git keeps them sorted for display)
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].Name < refs[j].Name
	})
	return refs, nil
}

// NameResolve resolves a given reference string to one or more equivalent hashes.
// Useful to:
//   - Convert a short hash to a list of matching full size hashes.
//   - Convert a symbolic, head or tag reference to a list of matching
//     full size hashes.
func (r *Repo) NameResolve(name string) ([]string, error) {
	matches := []string{}

	name = strings.TrimSpace(name)
	if name == "" {
		// Can't do much if nothing is given!
		return matches, nil
	}

	// Get all the references matching the given name if available
	refs, err := r.GetRefs(name, false)
	if err != nil {
		log.Printf("Can't resolve %s due to %v", name, err)
		return nil, err
	}

	// If HEAD is asked for, then get '.git/HEAD' as well.
	if name == "HEAD" {
		headHash, err := r.RefResolve("HEAD")
		if err == nil {
			log.Printf("Found HEAD reference as: %s\n", headHash)
			refs = append(refs, RefEntry{"HEAD", headHash})
		}
	}

	// The reference can be given as any part starting with last component in
	// a valid path. Such as "master", "heads/master" or "refs/heads/master", all
	// are valid references to a branch "master". The order of precendence is
	// as follows. The first matching entry is returned.
	//  HEAD
	// 	refs/<name>
	//  refs/tags/<name>
	//  refs/heads/<name>
	//  refs/remotes/<refname>
	//  refs/remotes/<refname>/HEAD
	//
	// Just pick the ref with shortest name as per the rules above.
	// TODO: enhance this algorithm with actual matches. Shortest way just
	// happens to work right now.
	if len(refs) > 0 {
		sort.Slice(refs, func(i, j int) bool {
			return len(refs[i].Name) < len(refs[j].Name)
		})
		log.Printf("Found %d references for %s", len(refs), name)
		matches = append(matches, refs[0].RefHash)
	}

	// The given name may even be a short or full hash.
	// Check if the given ref is a valid hexadecimal hash.
	re := regexp.MustCompile(`^[a-fA-F0-9]*$`)
	if !re.MatchString(name) {
		return matches, nil
	}

	// If the hash is smaller than 4, then return. "git" doesn't resolve
	// a hash smaller than 4 characters.
	// Also, git hashes are limited to 40 char (SHA1)
	if len(name) < 4 || len(name) > 40 {
		return matches, nil
	}

	// If the hash is given in full size, then use it as is.
	if len(name) == 40 {
		matches = append(matches, name)
		return matches, nil
	}

	// If reached here, then 'name' may be a valid short hash matching one or
	// more full hashes. Collect them all by looking at all files inside '.git/objects'.
	objectsPath, err := r.DirPath(true, "objects", name[0:2])
	if err != nil {
		log.Printf("Objects path not found for name %s (%v)", name, err)
		return matches, nil
	}

	// Read all files under this directory and collect all files matching the
	// remaining hash (after first 2 char).
	files, err := ioutil.ReadDir(objectsPath)
	if err != nil {
		log.Printf("Objects path dir %s access error (%v)", objectsPath, err)
		return matches, nil
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), name[2:]) {
			matches = append(matches, name[0:2]+file.Name())
		}
	}

	return matches, nil
}

// UniqueNameResolve converts a given name to a unique valid full object hash.
// It returns an error if there are less or more than 1 matching objects to the
// given name.
// This can be used by many commands to act on a single unique hash after taking
// a possible ambiguous name from the user.
func (r *Repo) UniqueNameResolve(name string) (string, error) {
	errmsg := fmt.Sprintf("fatal: ambiguous argument '%s': unknown revision or "+
		"path not in the working tree", name)

	matches, err := r.NameResolve(name)
	if err != nil || len(matches) == 0 {
		log.Printf("Failed to convert name %s to object hash or no matches "+
			"found: %v", name, err)
		return "", fmt.Errorf(errmsg)
	}

	// If there are more than one matches, then prepare an error with all the
	// matches.
	if len(matches) > 1 {
		msg := fmt.Sprintf("short SHA1 %s is ambiguous\n"+
			"Matching SHA1 list:\n"+
			"%s", name, strings.Join(matches, "\n"))
		return "", fmt.Errorf(msg)
	}

	return matches[0], nil
}

// ValidateRef strictly validates if a given reference is a valid reference
// in the local repository (or is HEAD). Returns the resolved object hash.
// This can be used by commands such as "gogit show-ref -verify".
func (r *Repo) ValidateRef(ref string) (string, error) {
	msg := "fatal: '{%s}' - not a valid ref"
	if ref != "HEAD" && !strings.HasPrefix(ref, "refs") {
		return "", fmt.Errorf(msg, ref)
	}

	refHash, err := r.UniqueNameResolve(ref)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf(msg, ref)
	}

	return refHash, nil
}
