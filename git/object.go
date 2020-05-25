package git

// ObjIntf is a common interface shared by all type of git objects.
type ObjIntf interface {
	Print() string
	Type() string
	DataSize() int
}

// Object is a struct holding the raw data for any git object type.
// ObjType can be one of "commit", "blob", "tree" or "tag".
type Object struct {
	ObjType string
	ObjData []byte
}

// NewObject returns a new git object of given type and with given data bytes.
func NewObject(objType string, data []byte) *Object {
	return &Object{
		ObjType: objType,
		ObjData: data,
	}
}
