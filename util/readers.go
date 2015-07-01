package util

import (
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func GetFromGithub(org, repo, branch, path, fileName string, w io.Writer) error {
	url := "https://rawgit.com/" + strings.Join([]string{org, repo, branch, path}, "/")
	w.Write([]byte("Will download from url -> " + url))
	return DownloadFromUrlToFile(url, fileName, w)
}

func GetFromIPFS(hash, fileName string, w io.Writer) error {
	url := IPFSBaseUrl() + hash
	w.Write([]byte("GETing file from IPFS. Hash -> " + hash + "\n"))
	return DownloadFromUrlToFile(url, fileName, w)
}

func DownloadFromUrlToFile(url, fileName string, w io.Writer) error {
	tokens := strings.Split(url, "/")
	if fileName == "" {
		fileName = tokens[len(tokens)-1]
	}
	w.Write([]byte("Downloading " + url + " to " + fileName + "\n"))

	output, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer output.Close()

	// adding manual timeouts as IPFS hangs for a while
	transport := http.Transport{
		Dial: dialTimeout,
	}
	client := http.Client{
		Transport: &transport,
	}
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func IPFSBaseUrl() string {
	var host string
	if os.Getenv("ERIS_CLI_CONTAINER") == "true" {
		host = "http://ipfs"
	} else {
		if os.Getenv("ERIS_IPFS_HOST") != "" {
			host = os.Getenv("ERIS_IPFS_HOST")
		} else {
			host = GetConfigValue("IpfsHost")
		}
	}
	return host + ":8080/ipfs/"
}

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
