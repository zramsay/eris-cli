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
	fmt.Println("Starting agent on localhost:17552")

	// cors.Default() sets up the middleware with default options being
	// all origins accepted with simple methods (GET, POST).
	// See https://github.com/rs/cors
	handler := cors.Default().Handler(mux)

	if err := http.ListenAndServe(":17552", handler); err != nil {
		return fmt.Errorf("Error starting agent: %v", err)
	}

	return nil
}

//func StopAgent(do *definitions.Do) error {
//	fmt.Println("Gracefully shutting down agent")
//	return nil
//}

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
		params, err := ParseDownloadURL(fmt.Sprintf("%s", r.URL))
		if err != nil {
			http.Error(w, `error parsing url`, http.StatusBadRequest)
			return
		}

		_, err = downloadBundle(params)
		if err != nil {
			// clean error handling...
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
		params, err := ParseInstallURL(fmt.Sprintf("%s", r.URL))
		if err != nil {
			http.Error(w, `error parsing url`, http.StatusBadRequest)
			return
		}

		// ensure chain to deploy on is running
		// might want to perform some other checks ... ?
		if !IsChainRunning(params["chainName"]) {
			http.Error(w, `chain name provided is not running`, http.StatusNotFound)
			return
		}

		installPath, err := downloadBundle(params)
		if err != nil {
			// clean error handling...
			return
		}
		//contractsPath := filepath.Join(installPath, params["dirName"])

		// user is authenticated
		// chain is running
		// contract bundle unbundled
		// time to deploy
		if err := DeployContractBundle(installPath, params["chainName"], params["address"]); err != nil {
			http.Error(w, `error deploying contract bundle`, http.StatusForbidden)
			// TODO reap bad addr error
			return
		}

		epmJSON := filepath.Join(installPath, "epm.json")
		epmByte, err := ioutil.ReadFile(epmJSON)
		if err != nil {
			http.Error(w, `error reading epm.json file`, http.StatusInternalServerError)
		}
		w.Write(epmByte)
	}
}

func downloadBundle(params map[string]string) (string, error) {
	bundleInfo := map[string]string{
		"groupId":  params["groupId"],
		"bundleId": params["bundleId"],
		"version":  params["version"],
		"hash":     params["hash"],
	}

	installPath := SetTarballPath(bundleInfo)

	tarBallPath, err := GetTarballFromIPFS(bundleInfo["hash"], installPath)
	if err != nil {
		//http.Error(w, `error getting from IPFS`, http.StatusInternalServerError)
		return "", err
	}

	if err = UnpackTarball(tarBallPath, installPath); err != nil {
		//http.Error(w, `error unpacking tarball`, http.StatusInternalServerError)
		return "", err
	}

	return installPath, nil
}

func ParseDownloadURL(url string) (map[string]string, error) {
	parsedURL, err := ParseURL(url)
	if err != nil {
		return nil, err
	}
	// TODO check fields

	return parsedURL, nil
}

func ParseInstallURL(url string) (map[string]string, error) {
	parsedURL, err := ParseURL(url)
	if err != nil {
		return nil, err
	}
	// TODO check fields

	return parsedURL, nil
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

	for _, p := range parsed {
		if p == "" {
			return nil, fmt.Errorf("empty field detected")
		}
	}

	return parsed, nil
}

func IsChainRunning(chainName string) bool {
	return util.IsChain(chainName, true)
}

// returns path to downloaded tarball
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

	doRun := definitions.NowDo()
	doRun.Path = path
	doRun.EPMConfigFile = filepath.Join(path, "epm.yaml")
	doRun.PackagePath = path
	doRun.ABIPath = path
	doRun.ChainName = chainName
	doRun.DefaultAddr = address

	if err := pkgs.RunPackage(doRun); err != nil {
		return err
	}
	log.Warn(doRun.Result)

	return nil
}

// deduplicate these two functions?
func SetTarballPath(bundleInfo map[string]string) string {
	groupID := strings.Replace(bundleInfo["groupId"], ".", "/", 1)
	bundleID := bundleInfo["bundleId"]
	version := bundleInfo["version"]

	return filepath.Join(common.BundlesPath, groupID, bundleID, version)
}

/*func SetBundlePath(bundleInfo map[string]string) string {
	groupID := strings.Replace(bundleInfo["groupId"], ".", "/", 1)
	bundleID := bundleInfo["bundleId"]
	version := bundleInfo["version"]
	dirName := bundleInfo["dirName"]

	return filepath.Join(common.BundlesPath, groupID, bundleID, version, dirName)
}*/

// ensure user is valid
// by querying the accounts chain
// TODO implement. Blocking on lack of details
//if !AuthenticateUser(params["user"]) {
//	w.Write([]byte(fmt.Sprintf("permissioned denied: %v\n", err)))
//	return
//}
