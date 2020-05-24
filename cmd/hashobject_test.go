package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

func TestHashObject(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "testGoGit")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	t.Logf("Testing hash-object with file: %s\n", tmpFile.Name())
	testStr := "Hello World\n"
	if _, err = tmpFile.Write([]byte(testStr)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	repo, err := git.GetRepo(".")
	if err != nil {
		t.Fatal(err)
	}

	blob, err := git.NewBlobFromFile(repo, tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// "want" is the output of 'echo "Hello World" | git hash-object --stdin'
	want := "557db03de997c86a4a028e1ebd3a1ceb225be238"
	got, err := repo.ObjectWrite(blob.Obj, false)
	util.Check(err)
	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, got, want)
}
