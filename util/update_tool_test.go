package util

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
			t.Fatalf("Expected git = %b and go = %b.\nResult: git = %b, go = %b", g.hasGit, g.hasGo, resultGit, resultGo)
		}
	}
}

func TestDownloadLatestRelease(t *testing.T) {
	filename, err := downloadLatestRelease()
	if err != nil {
		t.Fatal("Download failed with error:", err)
	} else {
		t.Log("Latest Release downloads successfully. Removing...")
		err := os.Remove(filename)
		if err != nil {
			t.Error("could not remove eris-cli binary test-download.", err)
		}
	}

}
