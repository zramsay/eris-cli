package update

import (
	"os"
	"testing"
)

type testCases struct {
	hasGit bool
	hasGo  bool
}

func TestCheckGitAndGo(t *testing.T) {
	var tests = []testCases{
		{true, true},
		{true, false},
		{false, true},
		{false, false},
	}
	for _, g := range tests {
		resultGit, resultGo := CheckGitAndGo(g.hasGit, g.hasGo)
		if g.hasGit != resultGit && g.hasGo != resultGo {
			t.Fatalf("Expected git = %v and go = %v.\nResult: git = %v, go = %v", g.hasGit, g.hasGo, resultGit, resultGo)
		}
	}
}

func TestDownloadLatestRelease(t *testing.T) {
	filename, err := DownloadLatestBinaryRelease()
	if err != nil {
		t.Fatal("Download failed with error:", err)
	} else {
		t.Log("Latest release downloads successfully. Removing")
		err := os.Remove(filename)
		if err != nil {
			t.Error("Could not remove Eris binary", err)
		}
	}

}
