package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient/external/github.com/docker/docker/pkg/archive"
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
	request.Close = true //for successive api calls
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
			return []byte(""), fmt.Errorf("error json unmarshaling body (bad request): %v", err)
		}
		return []byte(errs.Message), nil

		if response.StatusCode == http.StatusNotFound {
			if err = json.Unmarshal(body, &errs); err != nil {
				return []byte(""), fmt.Errorf("error json unmarshaling body (status not found): %v", err)
			}
			return []byte(errs.Message), nil
		}
	}
	//XXX hacky: would need to fix ipfs error msgs
	if string(body) == "Path Resolve error: context deadline exceeded" && string(body) == "context deadline exceeded" {
		return []byte(""), fmt.Errorf("A timeout occured while trying to reach IPFS. Run `eris files cache [hash], wait 5-10 seconds, then run `eris files [cmd] [hash]`")
	}
	return body, nil
}

func SendToIPFS(fileName string, bootstrap bool, w io.Writer) (string, error) {
	url := IPFSBaseGatewayUrl(bootstrap)
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
	url := IPFSBaseAPIUrl() + "pin/add?arg=" + fileHash

	w.Write([]byte("PINing file to IPFS. File =>\t" + fileHash + "\n"))
	body, err := PostAPICall(url, fileHash, w)
	if err != nil {
		return "", err
	}
	w.Write([]byte("Caching =>\t\t\t" + fileHash + ":" + url + "\n"))
	var p struct {
		Pinned []string
	}

	if err = json.Unmarshal(body, &p); err != nil {
		if fmt.Sprintf("%v", err) == "invalid character 'p' looking for beginning of value" {
			return "", fmt.Errorf("The file has already been pinned recusively (probably from ipfs add or eris files put). It is only possible to cache a file you don't already have. see issue #133 for more information")
		}
		return "", fmt.Errorf("unexpected error unmarshalling json: %v", err)
	}
	return p.Pinned[0], nil
}

func RemovePinnedFromIPFS(fileHash string, w io.Writer) (string, error) {
	url := IPFSBaseAPIUrl() + "pin/rm?arg=" + fileHash

	w.Write([]byte("Removing pinned file to IPFS. File =>\t" + fileHash + "\n"))
	body, err := PostAPICall(url, fileHash, w)
	if err != nil {
		return "", err
	}
	w.Write([]byte("Deleting from cache =>\t\t\t" + fileHash + ":" + url + "\n"))
	var p struct {
		Pinned []string
	}

	if err = json.Unmarshal(body, &p); err != nil {
		return "", fmt.Errorf("unexpected error unmarshalling json: %v", err)
	}

	return p.Pinned[0], nil
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
