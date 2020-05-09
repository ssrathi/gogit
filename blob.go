package gogit

type GitBlob struct {
	GitObject
}

func NewBlob(repo *Repo, data []byte) *GitBlob {
	blob := GitBlob{
		GitObject{
			Repository: repo,
			ObjType:    "blob",
			Data:       data,
		},
	}

	return &blob
}
