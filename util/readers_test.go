package util

import (
	"testing"
)

func TestTar(t *testing.T) {
	rc, err := Tar("balls", 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := Untar(rc, "shit", "fuck"); err != nil {
		t.Fatal(err)
	}

}
