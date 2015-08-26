package util

import (
	"io"
	"path/filepath"
	"strings"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient/external/github.com/docker/docker/pkg/archive"
)

//these were in writers.go but that got moved to /ipfs
func Tar(path string, compression archive.Compression) (io.ReadCloser, error) {
	return archive.Tar(path, compression)
}

func Untar(reader io.Reader, name, dest string) error {
	return archive.Untar(reader, dest, &archive.TarOptions{NoLchown: true, Name: name})
}

// note this function fails silently.
func GetGlobalLevelConfigFilesByType(typ string, withExt bool) []string {
	var path string
	switch typ {
	case "services":
		path = ServicesPath
	case "chains":
		path = BlockchainsPath
	case "actions":
		path = ActionsPath
	case "files":
		path = FilesPath
	}

	files := []string{}
	fileTypes := []string{}

	// TODO [csk]: DRY up how we deal with file extensions
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(path, t))
	}

	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			if !withExt {
				s1 = strings.Split(filepath.Base(s1), ".")[0]
			}
			files = append(files, s1)
		}
	}
	return files
}

func GetFileByNameAndType(typ, name string) string {
	logger.Debugf("Looking for file =>\t\t%s:%s\n", typ, name)
	files := GetGlobalLevelConfigFilesByType(typ, true)

	for _, file := range files {
		file = strings.Split(filepath.Base(file), ".")[0] // quick and dirty file root
		if file == name {
			logger.Debugf("This file found =>\t\t%s\n", file)
			return file
		}
		logger.Debugf("Group file found =>\t\t%s\n", file)
	}

	return ""
}
