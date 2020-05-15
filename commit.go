package gogit

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	AUTHOR_NAME  string = "Shyamsunder Rathi"
	AUTHOR_EMAIL string = "sxxxxxx@gmail.com"
)

type entryMap map[string][]string

type GitCommit struct {
	Obj     *GitObject
	Entries entryMap
	// Keep the keys to maintain the insertion order.
	Keys []string
	Msg  string
}

func NewCommit(obj *GitObject) (*GitCommit, error) {
	if obj.ObjType != "commit" {
		return nil, fmt.Errorf("Malformed object: bad type %s", obj.ObjType)
	}

	commit := GitCommit{
		Obj:     obj,
		Entries: entryMap{},
		Keys:    []string{},
	}

	// Parse the tree data.
	if err := commit.ParseData(); err != nil {
		return nil, err
	}

	return &commit, nil
}

// NewCommitFromParams builds a commit object using a 'tree' and optionall a
// 'parent' hash, and a given commit message.
// This can be used by CLI commands such as "gogit commit-tree".
func NewCommitFromParams(treeHash, parentHash, msg string) (*GitCommit, error) {
	data := []byte{}
	data = append(data, []byte("tree "+treeHash+"\n")...)
	if parentHash != "" {
		data = append(data, []byte("parent "+parentHash+"\n")...)
	}

	// Get the current time in <epoch zone-offset> format
	// Example: 1589530357 -0700
	cTime := time.Now()
	timeStamp := strconv.FormatInt(cTime.Unix(), 10) + " " + cTime.Format("-0700")

	// Build author and commiter values
	authorValue := fmt.Sprintf("%s <%s> %s", AUTHOR_NAME, AUTHOR_EMAIL, timeStamp)

	data = append(data, []byte("author "+authorValue+"\n")...)
	data = append(data, []byte("committer "+authorValue+"\n")...)
	data = append(data, byte('\n'))
	data = append(data, []byte(msg)...)

	obj := NewObject("commit", data)
	return NewCommit(obj)
}

func (commit *GitCommit) Type() string {
	return "commit"
}

func (commit *GitCommit) DataSize() int {
	return len(commit.Obj.ObjData)
}

func (commit *GitCommit) Print() string {
	var b strings.Builder

	// Print the key-values in insertion order first.
	for _, key := range commit.Keys {
		for _, val := range commit.Entries[key] {
			fmt.Fprintf(&b, "%s %s\n", key, val)
		}
	}

	// Print a blank line followed by the commit message.
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, commit.Msg)
	return b.String()
}

// ParseData parses a commit object's bytes and prepares a dictionary of its
// components.
func (commit *GitCommit) ParseData() error {
	/* Commit object has the following format:
	<key1> <value1>\n
	<key2> <value2 ...>\n
	 <value2 continued>\n
	 <value2 continued>\n
	<key3> <value3>\n
	...
	<blank line>
	<Remaining lines are part of commit-message. */
	datalen := len(commit.Obj.ObjData)
	for start := 0; start < datalen; {
		data := commit.Obj.ObjData[start:]

		spaceInd := bytes.IndexByte(data, byte(' '))
		newLenInd := bytes.IndexByte(data, byte('\n'))

		// Unless we have found a blank line, each line must have a space.
		// If the space is at first place, then it is part of the last value.
		// If the space is somewhere else, then it is a key-value pair.
		// Once a blank line is found, remaining lines are part of commit msg.
		if newLenInd < spaceInd {
			// Blank line, so remaining data is part of the commit msg.
			commit.Msg = string(data[1:])
			break
		}

		// Find the key which is the part before the space
		key := string(data[0:spaceInd])

		// The value can be single line or multi line.
		// Multi-line values have a space as the first character.
		var end int
		for {
			end = bytes.IndexByte(data, byte('\n'))
			if data[end+1] != byte(' ') {
				// This is not a continuation line, so stop!
				break
			}
		}

		// Get the value for this key and remove first space character on all
		// continuation lines.
		value := string(data[spaceInd+1 : end])
		value = strings.ReplaceAll(value, "\n ", "\n")

		// Save the key for insertion order if not already seen.
		// All keys with same values appear together in a commit msg.
		if _, ok := commit.Entries[key]; !ok {
			commit.Keys = append(commit.Keys, key)
		}

		// There can be multiple values for a single key.
		// Such as, there can be more than one 'parent' key for a commit.
		commit.Entries[key] = append(commit.Entries[key], value)

		// Move on to the next key-value pair.
		start += (end + 1)
	}

	return nil
}
