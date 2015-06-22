// parts of this file were abstracted from: https://github.com/mindreframer/golang-stuff/blob/master/github.com/dotcloud/docker/archive.go

package util

import (
	"io"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/docker/docker/pkg/archive"
)

func Tar(path string, compression archive.Compression) (io.ReadCloser, error) {
  return archive.Tar(path, compression)
}

func Untar(reader io.Reader, name, dest string) error {
  return archive.Untar(reader, dest, &archive.TarOptions{NoLchown: true, Name: name,})
}