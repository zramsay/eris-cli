package util

import (
	"encoding/json"
	"fmt"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient/vendor/github.com/docker/docker/pkg/archive"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func Tar(path string, compression archive.Compression) (io.ReadCloser, error) {
	return archive.Tar(path, compression)
}

func Untar(reader io.Reader, name, dest string) error {
	return archive.Untar(reader, dest, &archive.TarOptions{NoLchown: true, Name: name})
}

//returns []byte to let each command make its own struct for the response
//but handles the errs in here
func PostAPICall(url, fileHash string, w io.Writer) ([]byte, error) {
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return []byte(""), err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return []byte(""), err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte(""), err
	}

	var errs struct {
		Message string
		Code    int
	}
	if response.StatusCode >= http.StatusBadRequest {
		//TODO better err handling; this is a (very) slimed version of how IPFS does it.
		if err = json.Unmarshal(body, &errs); err != nil {
			return []byte(""), fmt.Errorf("error json unmarshaling body %v", err)
		}
		return []byte(errs.Message), nil

		if response.StatusCode == http.StatusNotFound {
			if err = json.Unmarshal(body, &errs); err != nil {
				return []byte(""), fmt.Errorf("error json unmarshaling body %v", err)
			}
			return []byte(errs.Message), nil
		}
	}
	return body, nil
}

func SendToIPFS(fileName string, w io.Writer) (string, error) {
	url := IPFSBaseUrl() + ":8080/ipfs/"
	w.Write([]byte("POSTing file to IPFS. File =>\t" + fileName + "\n"))
	head, err := UploadFromFileToUrl(url, fileName, w)
	if err != nil {
		return "", err
	}
	hash, ok := head["Ipfs-Hash"]
	if !ok || hash[0] == "" {
		return "", fmt.Errorf("No hash returned")
	}
	return hash[0], nil
}

func PinToIPFS(fileHash string, w io.Writer) (string, error) {
	url := IPFSBaseUrl() + ":5001/api/v0/pin/add?arg=" + fileHash

	w.Write([]byte("PINing file to IPFS. File =>\t" + fileHash + "\n"))
	body, err := PostAPICall(url, fileHash, w)
	if err != nil {
		return "", err
	}
	w.Write([]byte("Caching =>\t\t\t" + fileHash + ":" + url + "\n"))

	var p struct {
		Pinned []string
	}
	var msg string

	if err = json.Unmarshal(body, &p); err != nil {
		return "", fmt.Errorf("error json unmarshaling body %v", err)
	}
	msg = p.Pinned[0]
	return msg, nil
}

func UploadFromFileToUrl(url, fileName string, w io.Writer) (http.Header, error) {
	if url == "" {
		return http.Header{}, fmt.Errorf("To upload from a file to a url, I need a URL.")
	}
	w.Write([]byte("Uploading =>\t\t\t" + fileName + ":" + url + "\n"))

	input, err := os.Open(fileName)
	if err != nil {
		return http.Header{}, err
	}
	defer input.Close()

	request, err := http.NewRequest("POST", url, input)
	if err != nil {
		return http.Header{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return http.Header{}, err
	}
	defer response.Body.Close()

	return response.Header, nil
}
