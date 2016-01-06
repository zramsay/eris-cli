package util

import (
	"io"
	"strings"

	ipfs "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient/external/github.com/docker/docker/pkg/archive"
)

func Tar(path string, compression archive.Compression) (io.ReadCloser, error) {
	return archive.Tar(path, compression)
}

func Untar(reader io.Reader, name, dest string) error {
	return archive.Untar(reader, dest, &archive.TarOptions{NoLchown: true}) //, Name: name})
}

func GetFromGithub(org, repo, branch, path, directory, fileName string, w io.Writer) error {
	url := "https://raw.githubusercontent.com/" + strings.Join([]string{org, repo, branch, path}, "/")
	w.Write([]byte("Will download from url -> " + url))
	return ipfs.DownloadFromUrlToFile(url, fileName, directory, w)
}
