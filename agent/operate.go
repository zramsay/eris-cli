package agent

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/files"
	"github.com/eris-ltd/eris-cli/pkgs"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/Sirupsen/logrus"
	"github.com/eris-ltd/common/go/common"
	"github.com/rs/cors"
)

func StartAgent(do *definitions.Do) error {
	// unknown: auth details/https
	mux := http.NewServeMux()
	mux.HandleFunc("/chains", ListChains)
	mux.HandleFunc("/download", DownloadAgent)
	mux.HandleFunc("/install", InstallAgent)
	fmt.Println("Starting mux agent on localhost:17552")
	// cors.Default() sets up the middleware with default options being
	// all origins accepted with simple methods (GET, POST).
	// See https://github.com/rs/cors
	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":17552", handler)

	return nil
}

func StopAgent(do *definitions.Do) error {
	fmt.Println("Gracefully shutting down agent")
	return nil
}

func ListChains(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var deets []*util.Details

		deets = util.ErisContainersByType(definitions.TypeChain, true)
		names := make([]string, len(deets))
		for i, chainName := range deets {
			names[i] = chainName.ShortName
		}

		chainz := []string{}
		lenChainz := len(names) - 1
		for i, fmtName := range names {
			if i == 0 {
				chainz = append(chainz, "[")
			}
			if i == lenChainz {
				chainz = append(chainz, fmt.Sprintf("{ name: '%s' }]", fmtName))
				break
			}
			chainz = append(chainz, fmt.Sprintf("{ name: '%s' }, ", fmtName))
		}
		w.Write([]byte(strings.Join(chainz, "")))
	}
}

func DownloadAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Warn("Receiving request to download a contract bundle")
		params, err := ParseURL(fmt.Sprintf("%s", r.URL))
		if err != nil {
			http.Error(w, `error parsing url`, http.StatusBadRequest)
			return
		}

		bundleInfo := map[string]string{
			"groupId":  params["groupId"],
			"bundleId": params["bundleId"],
			"version":  params["version"],
		}

		for _, bI := range bundleInfo {
			if bI == "" {
				http.Error(w, `empty field detected`, http.StatusBadRequest)
				return
			}
		}

		installPath := SetTarballPath(bundleInfo)

		tarBallPath, err := GetTarballFromIPFS(params["hash"], installPath)
		if err != nil {
			http.Error(w, `error getting from IPFS`, http.StatusInternalServerError)
			return
		}

		if err = UnpackTarball(tarBallPath, installPath); err != nil {
			http.Error(w, `error unpacking tarball`, http.StatusInternalServerError)
			return
		}
	}
}

func InstallAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Warn("Receiving request to download and deploy a contract bundle")

		// parse response into various components
		// required to pull tarball, unpack on path
		// and deploy to running chain
		params, err := ParseURL(fmt.Sprintf("%s", r.URL))
		if err != nil {
			http.Error(w, `error parsing url`, http.StatusBadRequest)
			//w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte(fmt.Sprintf("error parsing url: %v\n", err)))
			return
		}

		bundleInfo := map[string]string{
			"groupId":  params["groupId"],
			"bundleId": params["bundleId"],
			"version":  params["version"],
		}

		for _, bI := range bundleInfo {
			if bI == "" {
				http.Error(w, `empty field detected`, http.StatusBadRequest)
				//w.WriteHeader(http.StatusBadRequest)
				//w.Write([]byte("empty field detected"))
				return
			}
		}

		installPath := SetTarballPath(bundleInfo)
		contractsPath := filepath.Join(installPath, params["dirName"])

		//TODO see bottom; implement User Authentication ...?

		// ensure chain to deploy on is running
		// might want to perform some other checks ... ?
		// true for chain that is running
		if IsChainRunning(params["chainName"]) {
			http.Error(w, `chain name provided is not running`, http.StatusNotFound)
			//w.WriteHeader(http.StatusNotFound)
			//w.Write([]byte(fmt.Sprintf("specified chain name is not running: %v\n", err)))
			return
		}

		// things are ready to go
		// let's get that tarball
		tarBallPath, err := GetTarballFromIPFS(params["hash"], installPath)
		if err != nil {
			http.Error(w, `error getting from IPFS`, http.StatusInternalServerError)
			//w.WriteHeader(http.StatusInternalServerError)
			//w.Write([]byte(fmt.Sprintf("error getting from IPFS: %v\n", err)))
			return
		}

		if err = UnpackTarball(tarBallPath, installPath); err != nil {
			http.Error(w, `error unpacking tarball`, http.StatusInternalServerError)
			//w.WriteHeader(http.StatusInternalServerError)
			//w.Write([]byte(fmt.Sprintf("error unpacking tarball: %v\n", err)))
			return
		}

		// user is authenticated
		// chain is running
		// contract bundle unbundled
		// time to deploy

		if err := DeployContractBundle(contractsPath, params["chainName"], params["address"]); err != nil {
			http.Error(w, `error deploying contract bundle`, http.StatusForbidden)
			// TODO reap bad addr error
			//w.WriteHeader(http.StatusForbidden)
			//w.Write([]byte(fmt.Sprintf("error deploying contract bundle: %v\n", err)))
			return
		}

		epmJSON := filepath.Join(contractsPath, "epm.json")
		epmByte, err := ioutil.ReadFile(epmJSON)
		if err != nil {
			http.Error(w, `error reading epm.json file`, http.StatusInternalServerError)
			//w.WriteHeader(http.StatusInternalServerError)
			//w.Write([]byte(fmt.Sprintf("error reading file: %v\n", err)))
		}
		w.Write(epmByte)
	}
}

// takes URL request for bundle install &
// returns a map of the things we need for installation
// assume it comes in as r.URL.Path[1:]
func ParseURL(url string) (map[string]string, error) {
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
	return util.IsChain(chainName, true)
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

	return filepath.Join(common.BundlesPath, groupID, bundleID, version)
}

func SetBundlePath(bundleInfo map[string]string) string {
	groupID := strings.Replace(bundleInfo["groupId"], ".", "/", 1)
	bundleID := bundleInfo["bundleId"]
	version := bundleInfo["version"]
	dirName := bundleInfo["dirName"]

	return filepath.Join(common.BundlesPath, groupID, bundleID, version, dirName)

}

// ensure user is valid
// by querying the accounts chain
// TODO implement. Blocking on lack of details
//if !AuthenticateUser(params["user"]) {
//	w.Write([]byte(fmt.Sprintf("permissioned denied: %v\n", err)))
//	return
//}
