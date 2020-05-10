package gogit

// A common interface shared by all type of git objects.
type GitType interface {
	Print() string
	Type() string
	DataSize() int
}

// A struct holding the raw data for any object type.
type GitObject struct {
	ObjType string
	ObjData []byte
}

func NewObject(objType string, data []byte) *GitObject {
	return &GitObject{
		ObjType: objType,
		ObjData: data,
	}
}
