package gogit

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strconv"
)

type GitObject struct {
	Repository *Repo
	ObjType    string
	Data       []byte
}

func NewObject(repo *Repo) *GitObject {
	return &GitObject{
		Repository: repo,
	}
}

// Find the data referred by the given sha1 hash and add the data to the
// object as per "Git" specifications.
func (obj *GitObject) Parse(objHash string) error {
	data_file, err := obj.Repository.FilePath(
		false, "objects", string(objHash[0:2]), string(objHash[2:]))
	if err != nil {
		return err
	}

	// Read the file data and decompress it.
	data, err := ioutil.ReadFile(data_file)
	if err != nil {
		return err
	}

	decompressed := bytes.NewBuffer(data)
	r, err := zlib.NewReader(decompressed)
	if err != nil {
		return fmt.Errorf("Malformed object %s: bad data", objHash)
	}
	data, _ = ioutil.ReadAll(r)
	r.Close()

	// Strip the header from the decompressed data.
	spaceInd := bytes.IndexByte(data, byte(' '))
	obj.ObjType = string(data[0:spaceInd])
	nullInd := bytes.IndexByte(data, byte('\x00'))

	size, err := strconv.Atoi(string(data[spaceInd+1 : nullInd]))
	if err != nil || len(data)-nullInd-1 != size {
		return fmt.Errorf("Malformed object %s: bad length", objHash)
	}

	// Save the header-stripped data.
	obj.Data = data[nullInd+1:]
	return nil
}

// Calculate the sha1 of a git object and optionally write it to a file as
// per "Git" specifications.
func (obj *GitObject) Write(write bool) (string, error) {
	// Prepare header
	header := []byte(obj.ObjType + " " + strconv.Itoa(len(obj.Data)) + "\x00")
	data := append(header, obj.Data...)

	// Compute sha1 of the bytes.
	h := sha1.New()
	h.Write(data)
	sha1hash := hex.EncodeToString(h.Sum(nil))

	if !write {
		return sha1hash, nil
	}

	// Write the compressed data to a path determined by the sha1 hash.
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(data)
	w.Close()

	data_file, err := obj.Repository.FilePath(
		true, "objects", string(sha1hash[0:2]), string(sha1hash[2:]))
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(data_file, compressed.Bytes(), 0444); err != nil {
		return "", err
	}

	// The data is now written to a file as per Git specification.

	return sha1hash, nil
}
