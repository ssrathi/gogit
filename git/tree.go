package git

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// TreeEntry is a struct holding components of a tree object.
type TreeEntry struct {
	mode    string
	hash    string
	objType string
	name    string
}

// Tree is a object with a list of "tree" entries and a git object.
type Tree struct {
	Repository *Repo
	*Object
	Entries []TreeEntry
}

// NewTree creates a new tree object by parsing a Object.
func NewTree(repo *Repo, obj *Object) (*Tree, error) {
	if obj.ObjType != "tree" {
		return nil, fmt.Errorf("Malformed object: bad type %s", obj.ObjType)
	}

	tree := Tree{
		Repository: repo,
		Object:     obj,
		Entries:    []TreeEntry{},
	}

	// Parse the tree data.
	if err := tree.ParseData(); err != nil {
		return nil, err
	}

	return &tree, nil
}

// NewTreeFromInput parses the given string and create a tree from it.
func NewTreeFromInput(repo *Repo, input string) (*Tree, error) {
	data := []byte{}
	entries := strings.Split(input, "\n")
	for _, entry := range entries {
		if len(entry) == 0 {
			continue
		}

		// Each entry is in the following format.
		// 100644 blob a1e5680b811ded8762390b94a40643293ee6c1b0<tab>README.md
		props := strings.Fields(entry)

		// Strip initial 0s from the mode if any.
		mode := strings.TrimLeft(props[0], "0")
		data = append(data, []byte(mode)...)
		data = append(data, byte(' '))
		data = append(data, []byte(props[3])...)
		data = append(data, byte('\x00'))

		// Append blob/tree hash as bytes.
		byteHash, err := hex.DecodeString(props[2])
		if err != nil {
			return nil, err
		}
		data = append(data, byteHash...)
	}

	obj := NewObject("tree", data)
	return NewTree(repo, obj)
}

// Type returns the type string of a tree object.
func (tree *Tree) Type() string {
	return "tree"
}

// DataSize returns the size of the data of a tree object.
func (tree *Tree) DataSize() int {
	return len(tree.ObjData)
}

// Print returns a string representation of a tree object.
func (tree *Tree) Print() string {
	var b strings.Builder
	for _, entry := range tree.Entries {
		// Prepend 0 in front of mode to make it 6 char long.
		entryMode := strings.Repeat("0", 6-len(entry.mode)) + entry.mode

		fmt.Fprintf(&b, "%s %s %s\t%s\n",
			entryMode, entry.objType, entry.hash, entry.name)
	}

	return b.String()
}

// ParseData parses a tree object's bytes and prepares a list of its components.
func (tree *Tree) ParseData() error {
	datalen := len(tree.ObjData)
	for start := 0; start < datalen; {
		// First get the mode which has a space after that.
		data := tree.ObjData[start:]
		spaceInd := bytes.IndexByte(data, byte(' '))
		entryMode := string(data[0:spaceInd])

		// Mode must be 40000 for directories and 100xxx for files.
		if len(entryMode) != 5 && len(entryMode) != 6 {
			return fmt.Errorf("Malformed object: bad mode %s", entryMode)
		}

		// Next get the name/path which has a null char after that.
		nameInd := bytes.IndexByte(data, byte('\x00'))
		entryName := string(data[spaceInd+1 : nameInd])

		// Next 20 bytes form the entry sha1 hash. It is in binary.
		entryHash := hex.EncodeToString(data[nameInd+1 : nameInd+21])

		// Get the type of each hash for printing.
		obj, err := tree.Repository.ObjectParse(entryHash)
		if err != nil {
			return err
		}

		// Prepare a new TreeEntry object and push it to the list.
		entry := TreeEntry{
			mode:    entryMode,
			hash:    entryHash,
			objType: obj.ObjType,
			name:    entryName,
		}
		tree.Entries = append(tree.Entries, entry)

		// Update the next starting point.
		start += (nameInd + 21)
	}

	// Sort the entries (git keeps them sorted for display)
	sort.Slice(tree.Entries, func(i, j int) bool {
		return tree.Entries[i].name < tree.Entries[j].name
	})

	return nil
}

// Checkout recreates an entire worktree in a given path by recursively reading
// the blobs and trees inside this tree object.
func (tree *Tree) Checkout(path string) error {
	for _, entry := range tree.Entries {
		createPath := filepath.Join(path, entry.name)
		obj, err := tree.Repository.ObjectParse(entry.hash)
		if err != nil {
			return err
		}

		if entry.objType == "tree" {
			// handle a tree object
			if err := os.Mkdir(createPath, os.ModePerm); err != nil {
				return err
			}

			// Recurse on the tree object.
			tree, err := NewTree(tree.Repository, obj)
			if err != nil {
				return err
			}

			err = tree.Checkout(createPath)
			if err != nil {
				return err
			}
		} else {
			// handle a blob object
			blob, err := NewBlob(tree.Repository, obj)
			if err != nil {
				return err
			}

			// Mode is in the format 100xxx. Strip the 100 from it.
			mode, _ := strconv.ParseInt(entry.mode[3:], 8, 32)
			err = ioutil.WriteFile(createPath, blob.ObjData, os.FileMode(mode))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
