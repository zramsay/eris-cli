package util

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monax/compilers/definitions"
)

var (
	ext string
)

// clear a directory of its contents
func ClearCache(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// Get language from filename extension
func LangFromFile(filename string) (string, error) {
	ext := path.Ext(filename)
	ext = strings.Trim(ext, ".")
	if _, ok := definitions.Languages[ext]; ok {
		return ext, nil
	}
	return "", unknownLang(ext)
}

// Unknown language error
func unknownLang(lang string) error {
	return fmt.Errorf("Unknown language %s", lang)
}

func CreateTemporaryFile(name string, code []byte) (*os.File, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	_, err = file.Write(code)
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	return file, nil
}
