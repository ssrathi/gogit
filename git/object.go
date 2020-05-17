package git

// GitType is a common interface shared by all type of git objects.
type GitType interface {
	Print() string
	Type() string
	DataSize() int
}

// GitObject is a struct holding the raw data for any object type.
type GitObject struct {
	ObjType string
	ObjData []byte
}

// NewObject returns a new object of given type and with given data bytes.
func NewObject(objType string, data []byte) *GitObject {
	return &GitObject{
		ObjType: objType,
		ObjData: data,
	}
}
