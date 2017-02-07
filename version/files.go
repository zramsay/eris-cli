// +build !arm

package version

var (
	SERVICE_DEFINITIONS = []string{
		"compilers",
		"ipfs",
		"keys",
		// used by [eris chains start myChain --logrotate]
		// but its docker image is not pulled
		"logrotate",
	}
)
