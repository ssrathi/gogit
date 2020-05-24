package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ssrathi/gogit/util"

	"github.com/ssrathi/gogit/git"
)

func assertEqual(t *testing.T, got interface{}, want interface{}) {
	t.Helper()

	if got != want {
		t.Fatalf("got '%+v' (%v), want '%+v' (%v)",
			got, reflect.TypeOf(got), want, reflect.TypeOf(want))
	}
}

func TestInit(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "testGoGit")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Logf("Testing init at dir: %s\n", tmpDir)
	_, err = git.NewRepo(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	gitDir := filepath.Join(tmpDir, ".git")
	assertEqual(t, util.IsPathPresent(gitDir), true)
	assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "HEAD")), true)
	assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "description")), true)
	assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "config")), true)
	assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "objects")), true)
	assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs")), true)
	assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs", "tags")), true)
	assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs", "heads")), true)
}

func TestInitNonEmpty(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "testGoGit")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Make this dir non-empty.
	_, err = ioutil.TempFile(tmpDir, "testfile")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Testing init at non-empty dir: %s\n", tmpDir)
	_, err = git.NewRepo(tmpDir)

	want := fmt.Sprintf("Work-tree %q is not empty", tmpDir)
	assertEqual(t, err.Error(), want)
}
