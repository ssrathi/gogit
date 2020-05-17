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

// TreeHash returns the "tree" object hash inside the given commit object.
// Each commit object has only one "tree" object inside it.
func (commit *GitCommit) TreeHash() string {
	return commit.Entries["tree"][0]
}

// Parents returns a list of parents of the given commit. If there are no
// parent (base commit), then it returns an empty list.
func (commit *GitCommit) Parents() []string {
	return commit.Entries["parent"]
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

// PrettyPrint prints a commit object in a human readable format, similar to
// what is shown by "git log" output.
func (commit *GitCommit) PrettyPrint() (string, error) {
	var b strings.Builder

	// Find the commit hash of this commit object first.
	repo, err := GetRepo(".")
	if err != nil {
		return "", err
	}
	commitHash, _ := repo.ObjectWrite(commit.Obj, false)

	// Print the needed key-values in "git log" format.
	fmt.Fprintf(&b, "commit %s\n", commitHash)
	authorEntry := commit.Entries["author"][0]
	// Author line is in the following format:
	// "<name1 name2 ...> <email> <epoch seconds> <timezone>"
	// Example: "Shyamsunder Rathi <sxxxxxx@gmail.com> 1589619289 -0700"
	items := strings.Fields(authorEntry)
	timezone := items[len(items)-1]
	epoch, _ := strconv.ParseInt(items[len(items)-2], 10, 64)
	epochTime := time.Unix(epoch, 0)
	// "git" time format in logs: "Sat May 16 19:26:38 2020 -0700"
	timeStr := epochTime.Format("Mon Jan 02 15:04:05 2006")

	email := items[len(items)-3]
	author := strings.Join(items[:len(items)-3], " ")
	fmt.Fprintf(&b, "Author: %s %s\n", author, email)
	fmt.Fprintf(&b, "Date:   %s %s\n", timeStr, timezone)

	// Print a blank line followed by the commit message.
	fmt.Fprintln(&b)

	// Message is printed with 4 lines indentation in each line.
	msgParts := strings.Split(commit.Msg, "\n")
	for i, msg := range msgParts {
		if len(msg) != 0 {
			fmt.Fprintf(&b, "    %s", msg)
		}
		if i != len(msgParts)-1 {
			fmt.Fprintln(&b)
		}
	}

	return b.String(), nil
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
		if spaceInd < 0 || newLenInd < spaceInd {
			// Blank line, so remaining data is part of the commit msg.
			commit.Msg = string(data[1:])
			break
		}

		// Find the key which is the part before the space
		key := string(data[0:spaceInd])

		// The value can be single line or multi line.
		// Multi-line values have a space as the first character.
		end := -1
		for {
			end += bytes.IndexByte(data[end+1:], byte('\n')) + 1
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
