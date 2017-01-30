package util

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
)

// used only for docker cp
func TarForDocker(pathToTar string, compression archive.Compression) (io.ReadCloser, error) {
	return archive.Tar(pathToTar, compression)
}

// or for just untarring :)
func UntarForDocker(reader io.Reader, name, dest string) error {
	return archive.Untar(reader, dest, &archive.TarOptions{NoLchown: true}) //, Name: name})
}

func PackTarball(pathToTar, nameOfTar string) (string, error) {
	reader, err := os.Open(pathToTar)
	if err != nil {
		return "", nil
	}

	fileName := filepath.Join(pathToTar, nameOfTar)

	writer, err := os.Create(fileName)
	if err != nil {
		return "", nil
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	if err != nil {
		return "", nil
	}

	return fileName, nil
}

// give a tarballs' path
// and the target installation directory
// for the ball in question
func UnpackTarball(tarBallPath, installPath string) error {
	// open tarball for reading
	reader, err := os.Open(tarBallPath)
	defer reader.Close()
	if err != nil {
		return fmt.Errorf("error opening %s: %v\n", tarBallPath, err)
	}

	return UntarForDocker(reader, "", installPath)
}

func GetFromGithub(org, repo, branch, p, directory, fileName string) error {
	url := "https://" + path.Join("raw.githubusercontent.com", org, repo, branch, p)
	return DownloadFromUrlToFile(url, fileName, directory)
}
