package ipfs

import (
	"fmt"
	"os"
)

var IpfsHost string = "http://0.0.0.0"

func IPFSBaseGatewayUrl(gateway string) string {
	if gateway == "eris" {
		return fmt.Sprintf("%s%s", sexyUrl(), ":8080/ipfs/")
	} else if gateway != "" {
		return fmt.Sprintf("%s%s", gateway, ":8080/ipfs/")
	} else {
		return fmt.Sprintf("%s%s", IPFSUrl(), ":8080/ipfs/")
	}
}

func IPFSBaseAPIUrl() string {
	return fmt.Sprintf("%s%s", IPFSUrl(), ":5001/api/v0/")
}

func sexyUrl() string {
	//TODO load balancer (when one isn't enough)
	return "http://ipfs.erisbootstrap.sexy"
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
