package agent

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/files"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/pkgs"
	"github.com/eris-ltd/eris-cli/util"
)

func checkHash(hash string) (string, error) {
	splitHash := strings.Split(hash, ".")
	if len(splitHash) >= 3 { // somefile.tar.gz for example
		lenTAR := len(splitHash) - 2
		lenGZ := len(splitHash) - 1
		//probably not ipfs hash & probably tar ball
		if splitHash[lenTAR] == "tar" && splitHash[lenGZ] == "gz" {
			log.Debug("hash provided appears to be a tarball")
			return "tarball", nil
		}
	}

	if len(hash) == 46 { // ipfs hash
		log.Debug("hash provided appears to be an IPFS hash")
		return "ipfs-hash", nil
	}
	return "", fmt.Errorf("unable to decipher ipfs hash provided")
}

func downloadBundleFromIPFS(params map[string]string) error {

	installPath := SetTarballPath(params)

	tarBallPath, err := GetTarballFromIPFS(params["hash"], installPath)
	if err != nil {
		return err
	}

	if err = UnpackTarball(tarBallPath, installPath); err != nil {
		return err
	}

	return nil
}

func downloadBundleFromTarball(body []byte, installPath, fileName string) error {
	if err := os.MkdirAll(installPath, 0777); err != nil {
		return err
	}

	tarBallPath := filepath.Join(installPath, fileName)
	if err := ioutil.WriteFile(tarBallPath, body, 0777); err != nil {
		return err
	}

	if err := UnpackTarball(tarBallPath, installPath); err != nil {
		return err
	}
	return nil
}

// takes URL request for bundle install &
// returns a map of the things we need for installation
// assume it comes in as r.URL.Path[1:]
func ParseURL(requiredArguments []string, url string) (map[string]string, error) {
	reqArgs := requiredArguments

	payload := strings.Split(url, "?")[1]
	infos := strings.Split(payload, "&")
	parsedURL := make(map[string]string, len(infos))

	for _, field := range infos {
		info := strings.Split(field, "=")
		parsedURL[info[0]] = info[1] // this is not very resilient
	}

	if len(parsedURL) != len(reqArgs) {
		return nil, fmt.Errorf("Wrong number of arguments:\n%v\nThese %v arguments are required:\n%s", parsedURL, len(reqArgs), strings.Join(reqArgs, ", "))
	}

	for _, arg := range reqArgs {
		if parsedURL[arg] == "" {
			return nil, fmt.Errorf("Missing field or bad argument name:\n%v\nThese fields cannot be empty:\n%s", parsedURL, strings.Join(reqArgs, ", "))
		}
	}

	return parsedURL, nil
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
	do.Hash = hash

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
	doRun.ABIPath = filepath.Join(path, "abi")
	doRun.ChainName = chainName
	doRun.DefaultAddr = address
	doRun.KeysPort = "4767"   // [csk] note this is too opinionated. down the road we should be reading from the service definition file to acquire right port
	doRun.ChainPort = "46657" // [csk] note this is too opinionated. down the road we should be reading from the chain definition file to acquire right port

	if err := pkgs.RunPackage(doRun); err != nil {
		return err
	}
	log.Warn(doRun.Result)

	return nil
}

func SetTarballPath(bundleInfo map[string]string) string {
	groupID := strings.Replace(bundleInfo["groupId"], ".", "/", 1)
	bundleID := bundleInfo["bundleId"]
	version := bundleInfo["version"]

	return filepath.Join(config.BundlesPath, groupID, bundleID, version)
}
