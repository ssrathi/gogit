package gogit

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type TreeEntry struct {
	mode    string
	hash    string
	objType string
	name    string
}

type GitTree struct {
	Obj     *GitObject
	Entries []TreeEntry
}

func NewTree(obj *GitObject) (*GitTree, error) {
	if obj.ObjType != "tree" {
		return nil, fmt.Errorf("Malformed object: bad type %s", obj.ObjType)
	}

	tree := GitTree{
		Obj:     obj,
		Entries: []TreeEntry{},
	}

	// Parse the tree data.
	if err := tree.ParseData(); err != nil {
		return nil, err
	}

	return &tree, nil
}

func (tree *GitTree) Type() string {
	return "tree"
}

func (tree *GitTree) DataSize() int {
	return len(tree.Obj.ObjData)
}

func (tree *GitTree) Print() string {
	var b strings.Builder
	for _, entry := range tree.Entries {
		fmt.Fprintf(&b, "%s %s %s\t%s\n",
			entry.mode, entry.objType, entry.hash, entry.name)
	}

	return b.String()
}

func (tree *GitTree) ParseData() error {
	repo, err := GetRepo(".")
	if err != nil {
		return err
	}

	datalen := len(tree.Obj.ObjData)
	for start := 0; start < datalen; {
		// First get the mode which has a space after that.
		data := tree.Obj.ObjData[start:]
		spaceInd := bytes.IndexByte(data, byte(' '))
		entryMode := string(data[0:spaceInd])

		// Mode must be 40000 for directories and 100xxx for files.
		if len(entryMode) != 5 && len(entryMode) != 6 {
			return fmt.Errorf("Malformed object: bad mode %s", entryMode)
		}

		// Prepend 0 in front of mode to make it 6 char long.
		entryMode = strings.Repeat("0", 6-len(entryMode)) + entryMode

		// Next get the name/path which has a null char after that.
		nameInd := bytes.IndexByte(data, byte('\x00'))
		entryName := string(data[spaceInd+1 : nameInd])

		// Next 20 bytes form the entry sha1 hash. It is in binary.
		entryHash := hex.EncodeToString(data[nameInd+1 : nameInd+21])

		// Get the type of each hash for printing.
		obj, err := repo.ObjectParse(entryHash)
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
