// parts of this file were abstracted from: https://github.com/mindreframer/golang-stuff/blob/master/github.com/dotcloud/docker/archive.go

package util

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/docker/docker/pkg/archive"
)

func Tar(path string, compression archive.Compression) (io.ReadCloser, error) {
	return archive.Tar(path, compression)
}

func Untar(reader io.Reader, name, dest string) error {
	return archive.Untar(reader, dest, &archive.TarOptions{NoLchown: true, Name: name})
}

func SendToIPFS(fileName string, w io.Writer) (string, error) {
	url := IPFSBaseUrl()
	w.Write([]byte("POSTing file to IPFS. File -> " + fileName + "\n"))
	return UploadFromFileToUrl(url, fileName, w)
}

func UploadFromFileToUrl(url, fileName string, w io.Writer) (string, error) {
	if url == "" {
		return "", fmt.Errorf("To upload from a file to a url, I need a URL.")
	}
	w.Write([]byte("Uploading " + fileName + " to " + url + "\n"))

	input, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer input.Close()

	request, err := http.NewRequest("POST", url, input)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	hash, ok := response.Header["Ipfs-Hash"]
	if !ok || hash[0] == "" {
		return "", fmt.Errorf("No hash returned")
	}

	return hash[0], nil
}
