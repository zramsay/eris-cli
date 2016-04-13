package ipfs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func SendToIPFS(fileName, gateway string, w io.Writer) (string, error) {
	url := IPFSBaseGatewayUrl(gateway)
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
		//XXX hacky
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
