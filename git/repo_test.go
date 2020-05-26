package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ssrathi/gogit/util"
)

var (
	repoDir    string
	repo       *Repo
	blob       *Blob
	tree       *Tree
	commit     *Commit
	testFile   string
	blobHash   string
	treeHash   string
	commitHash string
	testData   string
	treeInput  string
	commitMsg  string
)

// assertEqual checks if two given values are equal and fatals if not.
func assertEqual(t *testing.T, got interface{}, want interface{}) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%+[1]v' (%[1]T), want '%+[2]v' (%[2]T)", got, want)
	}
}

// setupTestArtifacts creates objects in a new repo for testing gogit commands.
func setupTestArtifacts() error {
	var err error
	repoDir, err = ioutil.TempDir(os.TempDir(), "testGoGit")
	if err != nil {
		return err
	}

	repo, err = NewRepo(repoDir)
	if err != nil {
		return err
	}
	testFile = "testfile"
	tmpFile, err := os.OpenFile(filepath.Join(repoDir, testFile),
		os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	testData = "Hello World\n"
	_, err = tmpFile.WriteString(testData)
	if err != nil {
		return err
	}
	tmpFile.Close()

	// Save this file's data as a blob first.
	blob, err = NewBlobFromFile(repo, tmpFile.Name())
	if err != nil {
		return err
	}
	blobHash, err = repo.ObjectWrite(blob.Object, true)
	if err != nil {
		return err
	}

	// Create a tree with this blob
	treeInput = fmt.Sprintf("100644 blob %s\t%s\n",
		blobHash, filepath.Base(tmpFile.Name()))
	tree, err = NewTreeFromInput(repo, treeInput)
	if err != nil {
		return err
	}

	// Write the tree now.
	treeHash, err = repo.ObjectWrite(tree.Object, true)
	if err != nil {
		return err
	}

	// Make a commit with this tree.
	commitMsg = "Test commit for testing\n"
	commit, err = NewCommitFromParams(repo, treeHash, "", commitMsg)
	if err != nil {
		return err
	}

	// Write the commit now.
	commitHash, err = repo.ObjectWrite(commit.Object, true)
	if err != nil {
		return err
	}

	// Update HEAD to point to this new commit.
	err = repo.UpdateRef("HEAD", commitHash)
	if err != nil {
		return err
	}

	return nil
}

func TestRepo(t *testing.T) {
	// Disable internal logs during test runs unless an ENV var is given.
	if os.Getenv("GOGIT_DBG") != "1" {
		log.SetOutput(ioutil.Discard)
		log.SetFlags(0)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Set up a git repo and create few objects in it for testing.
	err := setupTestArtifacts()
	defer os.RemoveAll(repoDir)
	assertEqual(t, err, nil)

	t.Logf("Repository directory: %s", repoDir)
	t.Logf("Blob  : %s", blobHash)
	t.Logf("Tree  : %s", treeHash)
	t.Logf("Commit: %s", commitHash)

	// Validate 'gogit' init operations by checking various folders and files
	// in the repo.
	t.Run("Validate repository", func(t *testing.T) {
		gitDir := filepath.Join(repoDir, ".git")
		assertEqual(t, util.IsPathPresent(gitDir), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "HEAD")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "description")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "config")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "objects")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs", "tags")), true)
		assertEqual(t, util.IsPathPresent(filepath.Join(gitDir, "refs", "heads")), true)
	})

	// Validate that a repo creation in a non-emptry dir fails.
	t.Run("Validate non-empty repository creation failure", func(t *testing.T) {
		_, err := NewRepo(repoDir)
		want := fmt.Sprintf("Work-tree %q is not empty", repoDir)
		assertEqual(t, err.Error(), want)
	})

	// Validate the blob hash from hash-object operation.
	// "want" is the output of 'echo "Hello World" | git hash-object --stdin'
	t.Run("Validate blob hash", func(t *testing.T) {
		got := blobHash
		want := "557db03de997c86a4a028e1ebd3a1ceb225be238"
		assertEqual(t, got, want)
	})

	// Validate the tree hash from mktree operation.
	// "want" is the output of 'git mktree' with input
	// '100644 blob 557db03de997c86a4a028e1ebd3a1ceb225be238	testfile'
	t.Run("Validate tree hash", func(t *testing.T) {
		got := treeHash
		want := "e592dfe791dd1e1cf202668707a5cfac07a635b3"
		assertEqual(t, got, want)
	})

	// Validate various cat-file options on a blob object.
	t.Run("Validate cat-file blob -p option", func(t *testing.T) {
		got := blob.Print()
		want := testData
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file blob -t option", func(t *testing.T) {
		got := blob.Type()
		want := "blob"
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file blob -s option", func(t *testing.T) {
		got := blob.DataSize()
		want := len(testData)
		assertEqual(t, got, want)
	})

	// Validate various cat-file options on a tree object.
	t.Run("Validate cat-file tree -p option", func(t *testing.T) {
		got := tree.Print()
		want := treeInput
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file tree -t option", func(t *testing.T) {
		got := tree.Type()
		want := "tree"
		assertEqual(t, got, want)
	})

	t.Run("Validate cat-file tree -s option", func(t *testing.T) {
		got := tree.DataSize()
		// "git cat-file -s" for this gives 36. This is not the same as length
		// of "treeInput" as the blob hash is stored as binary bytes.
		want := 36
		assertEqual(t, got, want)
	})

	// Validate various cat-file options on a commit object.
	t.Run("Validate cat-file commit -t option", func(t *testing.T) {
		got := commit.Type()
		want := "commit"
		assertEqual(t, got, want)
	})

	// Validate that the tree inside the commit matches the given tree hash.
	t.Run("Validate tree-hash inside commit", func(t *testing.T) {
		got := commit.TreeHash()
		want := treeHash
		assertEqual(t, got, want)
	})

	// Validate that the msg inside the commit matches the given message.
	t.Run("Validate message inside commit", func(t *testing.T) {
		got := commit.Msg
		want := commitMsg
		assertEqual(t, got, want)
	})

	// Validate that the parent inside the commit matches the given parent.
	t.Run("Validate parent inside commit", func(t *testing.T) {
		parents := commit.Parents()
		// No parent is given to a base commit.
		assertEqual(t, len(parents), 0)
	})

	// Validate that the author details inside the commit matches the given values.
	t.Run("Validate author inside commit", func(t *testing.T) {
		name, email := commit.Author()
		assertEqual(t, name, AuthorName)
		assertEqual(t, email, AuthorEmail)
	})

	// Validate that a blob can be parsed from a given blob-hash.
	t.Run("Validate blob creation with a blob-hash", func(t *testing.T) {
		obj, err := repo.ObjectParse(blobHash)
		assertEqual(t, err, nil)

		testBlob, err := NewBlob(repo, obj)
		assertEqual(t, err, nil)

		assertEqual(t, testBlob.Print(), testData)
		assertEqual(t, testBlob.DataSize(), len(testData))
	})

	// Validate that a tree can be parsed from a given tree-hash.
	t.Run("Validate tree creation with a tree-hash", func(t *testing.T) {
		obj, err := repo.ObjectParse(treeHash)
		assertEqual(t, err, nil)

		testTree, err := NewTree(repo, obj)
		assertEqual(t, err, nil)

		// "git cat-file -s" for this gives 36. This is not the same as length
		// of "treeInput" as the blob hash is stored as binary bytes.
		assertEqual(t, testTree.Print(), treeInput)
		assertEqual(t, testTree.DataSize(), 36)
	})

	// Validate that a commit can be parsed from a given commit-hash.
	t.Run("Validate commit creation with a commit-hash", func(t *testing.T) {
		obj, err := repo.ObjectParse(commitHash)
		assertEqual(t, err, nil)

		testCommit, err := NewCommit(repo, obj)
		assertEqual(t, err, nil)

		assertEqual(t, testCommit.Msg, commitMsg)
		assertEqual(t, testCommit.TreeHash(), treeHash)

		name, email := testCommit.Author()
		assertEqual(t, name, AuthorName)
		assertEqual(t, email, AuthorEmail)
	})

	// Validate various rev-parse arguments.
	t.Run("Validate rev-parse HEAD", func(t *testing.T) {
		for _, revision := range []string{
			"HEAD",
			commitHash[:4],
			commitHash[:7],
			commitHash[:10],
			commitHash[:20],
			commitHash,
			"master",
			"heads/master",
			"refs/heads/master",
		} {
			objHash, err := repo.UniqueNameResolve(revision)
			assertEqual(t, err, nil)
			assertEqual(t, objHash, commitHash)
		}
	})

	t.Run("Validate rev-parse less than four short hash", func(t *testing.T) {
		_, err := repo.UniqueNameResolve(commitHash[:3])
		want := fmt.Errorf("fatal: ambiguous argument '%s': unknown revision "+
			"or path not in the working tree", commitHash[:3])
		assertEqual(t, err, want)
	})

	t.Run("Validate rev-parse invalid", func(t *testing.T) {
		_, err := repo.UniqueNameResolve("FOO")
		want := fmt.Errorf("fatal: ambiguous argument 'FOO': unknown revision " +
			"or path not in the working tree")
		assertEqual(t, err, want)
	})

	// Validate 'show-ref' outputs.
	t.Run("Validate show-ref", func(t *testing.T) {
		refs, err := repo.GetRefs("", false /* showHead */)
		assertEqual(t, err, nil)
		assertEqual(t, len(refs), 1)
		assertEqual(t, refs[0].RefHash, commitHash)
		assertEqual(t, refs[0].Name, "refs/heads/master")
	})

	t.Run("Validate show-ref with HEAD", func(t *testing.T) {
		refs, err := repo.GetRefs("", true /* showHead */)
		assertEqual(t, err, nil)
		assertEqual(t, len(refs), 2)

		assertEqual(t, refs[0].RefHash, commitHash)
		assertEqual(t, refs[0].Name, "HEAD")

		assertEqual(t, refs[1].RefHash, commitHash)
		assertEqual(t, refs[1].Name, "refs/heads/master")
	})

	// Validate 'checkout' functionality.
	t.Run("Validate checkout", func(t *testing.T) {
		checkoutDir, err := ioutil.TempDir(os.TempDir(), "testGoGitCheckout")
		assertEqual(t, err, nil)
		defer os.RemoveAll(checkoutDir)

		t.Logf("Checkout test directory: %s", checkoutDir)
		err = tree.Checkout(checkoutDir)
		assertEqual(t, err, nil)

		// testFile should be created in this new path with original data.
		dataFile := filepath.Join(checkoutDir, testFile)
		assertEqual(t, util.IsPathPresent(dataFile), true)
		data, err := ioutil.ReadFile(dataFile)
		assertEqual(t, err, nil)
		assertEqual(t, string(data), testData)
	})

	// Validate 'update-ref' functionality.
	t.Run("Validate update-ref HEAD to master", func(t *testing.T) {
		for _, newValue := range []string{
			"master",
			"heads/master",
			"refs/heads/master",
		} {
			err := repo.UpdateRef("HEAD", newValue)
			assertEqual(t, err, nil)
			masterHash, err := repo.UniqueNameResolve("HEAD")
			assertEqual(t, err, nil)
			assertEqual(t, masterHash, commitHash)
		}
	})

	t.Run("Validate update-ref a new branch to HEAD", func(t *testing.T) {
		ref := "refs/heads/new_branch"
		err := repo.UpdateRef(ref, "HEAD")
		assertEqual(t, err, nil)
		masterHash, err := repo.UniqueNameResolve(ref)
		assertEqual(t, err, nil)
		assertEqual(t, masterHash, commitHash)
	})

	t.Run("Validate update-ref with an invalid ref", func(t *testing.T) {
		newValue := "refs/heads/non-existent-ref"
		err := repo.UpdateRef("HEAD", newValue)
		want := fmt.Errorf("fatal: '{%s}' - not a valid SHA1", newValue)
		assertEqual(t, err, want)
	})
}
