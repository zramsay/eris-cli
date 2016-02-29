package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	dir              = "idi"
	dirToTar         = filepath.Join(os.TempDir(), dir)
	tarBallName      = "war.tar.gz"
	testFile         = "general.txt"
	testFileContents = "was not a marmot"
	installPath      = filepath.Join(os.TempDir(), "not_idi")
)

func testPackTarball(t *testing.T) string {
	// write file in in dir that'll be tarred
	// contents of file to be tested
	if err := os.MkdirAll(dirToTar, 0777); err != nil {
		t.Fatalf("error making dir: %v\n", err)
	}

	file, err := ioutil.TempFile(dirToTar, testFile)
	if err != nil {
		// XXX fails hmf
		t.Fatalf("error creating tempfile: %v\n", err)
	}
	defer file.Close()
	defer os.Remove(file.Name())

	_, err = file.Write([]byte(testFileContents))
	if err != nil {
		t.Fatalf("error writing file: %v\n", err)
	}

	//write the file
	pathOfBall, err := PackTarball(dirToTar, tarBallName)
	if err != nil {
		t.Fatalf("error packing tarball: %v\n", err)
	}
	return pathOfBall

}

func TestUnpackTarball(t *testing.T) {
	pathOfBall := testPackTarball(t)

	if err := UnpackTarball(pathOfBall, installPath); err != nil {
		t.Fatalf("error unpacking tarball: %v", err)
	}

	file, err := ioutil.ReadFile(filepath.Join(installPath, dir, testFile))
	if err != nil {
		t.Fatalf("err reading file: %v\n", err)
	}

	if testFileContents != string(file) {
		t.Fatalf("contents do not match. expected (%s) got (%s)", testFileContents, string(file))
	}

	//read contents of installPath
}
