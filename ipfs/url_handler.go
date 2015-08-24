package ipfs

import (
	"os"

	"github.com/eris-ltd/eris-cli/util"
)

//XXX url funcs can take flags for which host to go to.
func IPFSBaseGatewayUrl(bootstrap bool) string {
	if bootstrap {
		return sexyUrl() + ":8080/ipfs/"
	} else {
		return IPFSUrl() + ":8080/ipfs/"
	}
}

func IPFSBaseAPIUrl() string {
	return IPFSUrl() + ":5001/api/v0/"
}

func sexyUrl() string {
	//bootstrap was down
	//TODO fix before merge; DNS + load balancer
	return "http://147.75.194.73"
}

func IPFSUrl() string {
	var host string
	if os.Getenv("ERIS_CLI_CONTAINER") == "true" {
		host = "http://ipfs"
	} else {
		if os.Getenv("ERIS_IPFS_HOST") != "" {
			host = os.Getenv("ERIS_IPFS_HOST")
		} else {
			host = util.GetConfigValue("IpfsHost")
		}
	}
	return host
}
