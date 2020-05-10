package gogit

import (
	"fmt"
	"io/ioutil"
)

type GitBlob struct {
	Obj *GitObject
}

func NewBlob(obj *GitObject) (*GitBlob, error) {
	if obj.ObjType != "blob" {
		return nil, fmt.Errorf("Malformed object: bad type %s", obj.ObjType)
	}

	blob := GitBlob{
		Obj: obj,
	}
	return &blob, nil
}

func NewBlobFromFile(file string) (*GitBlob, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	blob := GitBlob{
		Obj: NewObject("blob", data),
	}
	return &blob, nil
}

func (blob *GitBlob) Print() string {
	return string(blob.Obj.ObjData)
}

func (blob *GitBlob) Type() string {
	return "blob"
}

func (blob *GitBlob) DataSize() int {
	return len(blob.Obj.ObjData)
}
