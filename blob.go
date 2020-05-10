package gogit

import (
	"io/ioutil"
)

type GitBlob struct {
	GitObject
}

func NewBlob(repo *Repo, file string) (*GitBlob, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	blob := GitBlob{
		GitObject{
			Repository: repo,
			ObjType:    "blob",
			Data:       data,
		},
	}

	return &blob, nil
}
