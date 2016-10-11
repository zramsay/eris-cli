package ipfs

import (
	"fmt"
	"os"
)

var IpfsHost string = "http://0.0.0.0"
var IpfsPort string = "8080"

func IPFSBaseGatewayUrl(gateway, port string) string {
	if port == "" {
		port = IpfsPort
	}
	if gateway == "eris" {
		return fmt.Sprintf("%s:%s%s", SexyUrl(), port, "/ipfs/")
	} else if gateway != "" {
		return fmt.Sprintf("%s:%s%s", gateway, port, "/ipfs/")
	} else {
		return fmt.Sprintf("%s:%s%s", IPFSUrl(), port, "/ipfs/")
	}
}

func IPFSBaseAPIUrl() string {
	return fmt.Sprintf("%s%s", IPFSUrl(), ":5001/api/v0/")
}

func SexyUrl() string {
	//TODO load balancer (when one isn't enough)
	return "http://ipfs.monax.io"
}

func IPFSUrl() string {
	var host string
	if os.Getenv("ERIS_CLI_CONTAINER") == "true" {
		host = "http://ipfs"
	} else {
		if os.Getenv("ERIS_IPFS_HOST") != "" {
			host = os.Getenv("ERIS_IPFS_HOST")
		} else {
			host = IpfsHost
		}
	}
	return host
}
