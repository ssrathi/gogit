package git

import (
	"fmt"
	"io/ioutil"
)

// GitBlob is a git object to represent the data of a single file.
type GitBlob struct {
	Repository *Repo
	Obj        *GitObject
}

// NewBlob creates a new blob object by parsing a GitObject.
func NewBlob(repo *Repo, obj *GitObject) (*GitBlob, error) {
	if obj.ObjType != "blob" {
		return nil, fmt.Errorf("Malformed object: bad type %s", obj.ObjType)
	}

	blob := GitBlob{
		Repository: repo,
		Obj:        obj,
	}
	return &blob, nil
}

// NewBlobFromFile creates a new blob object by reading data from a file.
func NewBlobFromFile(repo *Repo, file string) (*GitBlob, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	blob := GitBlob{
		Repository: repo,
		Obj:        NewObject("blob", data),
	}
	return &blob, nil
}

// Print returns a string representation of a blob object.
func (blob *GitBlob) Print() string {
	return string(blob.Obj.ObjData)
}

// Type returns the type string of a blob object.
func (blob *GitBlob) Type() string {
	return "blob"
}

// DataSize returns the size of the data of a blob object.
func (blob *GitBlob) DataSize() int {
	return len(blob.Obj.ObjData)
}
