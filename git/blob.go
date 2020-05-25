package git

import (
	"fmt"
	"io/ioutil"
)

// Blob is a git object to represent the data of a single file.
type Blob struct {
	Repository *Repo
	*Object
}

// NewBlob creates a new blob object by parsing a Object.
func NewBlob(repo *Repo, obj *Object) (*Blob, error) {
	if obj.ObjType != "blob" {
		return nil, fmt.Errorf("Malformed object: bad type %s", obj.ObjType)
	}

	blob := Blob{
		Repository: repo,
		Object:     obj,
	}
	return &blob, nil
}

// NewBlobFromFile creates a new blob object by reading data from a file.
func NewBlobFromFile(repo *Repo, file string) (*Blob, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	blob := Blob{
		Repository: repo,
		Object:     NewObject("blob", data),
	}
	return &blob, nil
}

// Print returns a string representation of a blob object.
func (blob *Blob) Print() string {
	return string(blob.ObjData)
}

// Type returns the type string of a blob object.
func (blob *Blob) Type() string {
	return "blob"
}

// DataSize returns the size of the data of a blob object.
func (blob *Blob) DataSize() int {
	return len(blob.ObjData)
}
