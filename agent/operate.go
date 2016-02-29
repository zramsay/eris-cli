package agent

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/files"
	"github.com/eris-ltd/eris-cli/pkgs"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func StartAgent(do *definitions.Do) error {
	// unknown: auth details/https
	mux := http.NewServeMux()
	mux.HandleFunc("/install", InstallAgent)
	fmt.Println("Starting mux agent on localhost:17552")
	http.ListenAndServe(":17552", mux)

	return nil
}

func StopAgent(do *definitions.Do) error {
	fmt.Println("Gracefully shutting down agent")
	return nil
}

func InstallAgent(w http.ResponseWriter, r *http.Request) {
	// TODO define an error type or something
	if r.Method == "POST" {
		log.Warn("Receiving request to install contract bundle")

		// parse response into various components
		// required to pull tarball, unpack on path
		// and deploy to running chain
		params, err := ParseInstallURL(fmt.Sprintf("%s", r.URL))
		if err != nil {
			w.Write([]byte(fmt.Sprintf("error parsing url: %v\n", err)))
			return
		}

		bundleInfo := map[string]string{
			"groupId":  params["groupId"],
			"bundleId": params["bundleId"],
			"version":  params["version"],
			//"dirName":  params["dirName"],
		}

		installPath := SetTarballPath(bundleInfo)
		contractsPath := filepath.Join(installPath, params["dirName"])

		//TODO see bottom; implement User Authentication ...?

		// ensure chain to deploy on is running
		// might want to perform some other checks ... ?
		if !IsChainRunning(params["chainName"]) {
			w.Write([]byte(fmt.Sprintf("chainName is not running: %v\n", err)))
			return
		}

		// things are ready to go
		// let's get that tarball
		tarBallPath, err := GetTarballFromIPFS(params["hash"], installPath)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("error getting from IPFS: %v\n", err)))
			return
		}

		// return something...?
		if err = UnpackTarball(tarBallPath, installPath); err != nil {
			w.Write([]byte(fmt.Sprintf("error unpacking tarball: %v\n", err)))
			return
		}

		// user is authenticated
		// chain is running
		// contract bundle unbundled
		// time to deploy

		if err := DeployContractBundle(contractsPath, params["chainName"], params["address"]); err != nil {
			w.Write([]byte(fmt.Sprintf("error deploying contract bundle: %v\n", err)))
		}
	}
}

// takes URL request for bundle install &
// returns a map of the things we need for installation
// assume it comes in as r.URL.Path[1:]
func ParseInstallURL(url string) (map[string]string, error) {
	payload := strings.Split(url, "?")[1]
	infos := strings.Split(payload, "&")
	parsed := make(map[string]string, len(infos))

	for _, field := range infos {
		info := strings.Split(field, "=")
		parsed[info[0]] = info[1] // this is not very resilient
	}

	return parsed, nil
}

// check that user making request is authenticated
// on the accounts chain.
//TODO implement
func AuthenticateUser(user string) bool {
	var yes bool
	if user == "sire" { // for test
		yes = true
	} else { // temp
		yes = true
	}
	return yes
}

func IsChainRunning(chainName string) bool {
	doCh := definitions.BlankChain()
	doCh.Name = chainName
	return chains.IsChainRunning(doCh)
}

func GetTarballFromIPFS(hash, installPath string) (string, error) {
	if !util.DoesDirExist(installPath) {
		if err := os.MkdirAll(installPath, 0777); err != nil {
			return "", err
		}
	}
	do := definitions.NowDo()
	do.Name = hash

	do.Path = filepath.Join(installPath, fmt.Sprintf("%s%s", hash, ".tar.gz"))
	if err := files.GetFiles(do); err != nil {
		return "error", err
	}

	return do.Path, nil
}

// for consistency in InstallAgent()
func UnpackTarball(tarBallPath, installPath string) error {
	return util.UnpackTarball(tarBallPath, installPath)
}

func DeployContractBundle(path, chainName, address string) error {
	//need to get address from keys container
	//or have it provided somewhere

	//os.Chdir(path)

	doRun := definitions.NowDo()
	doRun.Path = path
	doRun.EPMConfigFile = filepath.Join(path, "epm.yaml")
	doRun.PackagePath = path
	doRun.ABIPath = path
	doRun.ChainName = chainName
	doRun.DefaultAddr = address
	//doRun.Debug = true

	if err := pkgs.RunPackage(doRun); err != nil {
		log.Warn(doRun.Result)
		fmt.Println(doRun.Result)
		return err
	}
	log.Warn(doRun.Result)

	return nil
}

func SetTarballPath(bundleInfo map[string]string) string {
	groupID := strings.Replace(bundleInfo["groupId"], ".", "/", 1)
	bundleID := bundleInfo["bundleId"]
	version := bundleInfo["version"]
	//dirName := bundleInfo["dirName"]

	return filepath.Join(common.BundlesPath, groupID, bundleID, version, "auth") //dirName) //not sure why "auth" => from slide deck
}

func SetBundlePath(bundleInfo map[string]string) string {
	groupID := strings.Replace(bundleInfo["groupId"], ".", "/", 1)
	bundleID := bundleInfo["bundleId"]
	version := bundleInfo["version"]
	dirName := bundleInfo["dirName"]

	return filepath.Join(common.BundlesPath, groupID, bundleID, version, "auth", dirName) //not sure why "auth" => from slide deck

}

// ensure user is valid
// by querying the accounts chain
// TODO implement. Blocking on lack of details
//if !AuthenticateUser(params["user"]) {
//	w.Write([]byte(fmt.Sprintf("permissioned denied: %v\n", err)))
//	return
//}
