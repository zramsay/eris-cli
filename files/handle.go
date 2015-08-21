package files

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

func GetFiles(do *definitions.Do) error {
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")
	if do.CSV != "" {
		logger.Debugf("Gonna Import the files from =>\t\t%s into %v\n", do.CSV, do.NewName)
		err = importFiles(do.CSV, do.NewName)

	} else {
		logger.Debugf("Gonna Import a file =>\t\t%s:%v\n", do.Name, do.Path)
		err = importFile(do.Name, do.Path)
	}
	if err != nil {
		return err
	}
	do.Result = "success"
	return nil
}

func PutFiles(do *definitions.Do) error {
	var hash string
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")

	if do.AddDir {
		logger.Debugf("Gonna add the contents of a directory =>\t\t%s:%v\n", do.Name, do.Path)
		hashes, err := exportDir(do.Name, do.Gateway)
		if err != nil {
			return err
		}
		do.Result = hashes
	} else {
		logger.Debugf("Gonna Add a file =>\t\t%s:%v\n", do.Name, do.Path)
		hash, err = exportFile(do.Name, do.Gateway)
		if err != nil {
			return err
		}
		do.Result = hash
	}
	//make string flag that defaults to sexy but can point anywhere
	if do.Gateway {
		logger.Debugf("Posting to ipfs.erisbootstrap.sexy")
	} else {
		logger.Debugf("Posting to gateway.ipfs.io")
	}

	return nil
}

func PinFiles(do *definitions.Do) error {
	var hash string
	var hashes string
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")

	if do.CSV != "" {
		logger.Debugf("Gonna Pin all the files from =>\t\t%s\n", do.CSV)
		hashes, err = pinFiles(do.CSV)
		do.Result = hashes

	} else {
		logger.Debugf("Gonna Pin a file =>\t\t%s:%v\n", do.Name, do.Path)
		hash, err = pinFile(do.Name)
		do.Result = hash
	}
	if err != nil {
		return err
	}
	return nil
}

func CatFiles(do *definitions.Do) error {
	var hash string
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")
	logger.Debugf("Gonna Cat a file =>\t\t%s:%v\n", do.Name, do.Path)
	hash, err = catFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func ListFiles(do *definitions.Do) error {
	var hash string
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")
	logger.Debugf("Gonna List an object =>\t\t%s:%v\n", do.Name, do.Path)
	hash, err = listFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func ListPinned(do *definitions.Do) error {
	var hash string
	var hashes string
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")

	if do.Rm {
		logger.Infoln("Removing all cached files")
		hashes, err = rmAllPinned()
		do.Result = hashes
	} else if do.Hash != "" {
		logger.Infof("Removing %v, from cache", do.Hash)
		hashes, err = rmPinnedByHash(do.Hash)
		do.Result = hashes
	} else {
		logger.Debugf("Listing files pinned locally")
		hash, err = listPinned()
		do.Result = hash
	}
	if err != nil {
		return err
	}
	return nil
}

func importFile(hash, fileName string) error {
	var err error

	if logger.Level > 0 {
		err = util.GetFromIPFS(hash, fileName, "", logger.Writer)
	} else {
		err = util.GetFromIPFS(hash, fileName, "", bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return err
	}
	return nil
}

func importFiles(csvfile, newdir string) error {
	var err error

	csvFile, err := os.Open(csvfile)
	if err != nil {
		return fmt.Errorf("error opening csv file: %v\n", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading csv file: %v\n", err)
	}
	//TODO mkdir (dirname) and dump files in there rather than in pwd
	//file outputs tied up in GetFromIPFS
	for _, each := range rawCSVdata {
		if logger.Level > 0 {
			err = util.GetFromIPFS(each[0], each[1], newdir, logger.Writer)
		} else {
			err = util.GetFromIPFS(each[0], each[1], newdir, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func exportFile(fileName string, gateway bool) (string, error) {
	var hash string
	var err error

	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, gateway, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, gateway, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}

	return hash, nil
}

func exportDir(dirName string, gateway bool) (string, error) {
	var hashes string
	var err error

	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return "", fmt.Errorf("error reading directory %v\n", err)
	}
	gwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory %v\n", err)
	}
	hashArray := make([]string, len(files))
	fileNames := make([]string, len(files))
	//the dir ends up in the loop & tries to post
	for i := range files {
		//hacky
		file := path.Join(gwd, dirName, files[i].Name())
		if logger.Level > 0 {
			hashArray[i], err = util.SendToIPFS(file, gateway, logger.Writer)
		} else {
			hashArray[i], err = util.SendToIPFS(file, gateway, bytes.NewBuffer([]byte{}))
		}
		if err != nil {
			return "", fmt.Errorf("error reading file %v\n", err)
		}
		fileNames[i] = files[i].Name()
	}

	err = writeCsv(hashArray, fileNames)
	if err != nil {
		return "", err
	}

	hashes = strings.Join(hashArray, "\n")

	return hashes, nil
}

func pinFile(fileHash string) (string, error) {
	var hash string
	var err error

	if logger.Level > 0 {
		hash, err = util.PinToIPFS(fileHash, logger.Writer)
	} else {
		hash, err = util.PinToIPFS(fileHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func pinFiles(csvfile string) (string, error) {
	var err error

	csvFile, err := os.Open(csvfile)
	if err != nil {
		return "", fmt.Errorf("error opening csv file: %v\n", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("error reading csv file: %v\n", err)
	}

	hashArray := make([]string, len(rawCSVdata))
	for i, each := range rawCSVdata {
		if logger.Level > 0 {
			hashArray[i], err = util.PinToIPFS(each[0], logger.Writer)
		} else {
			hashArray[i], err = util.PinToIPFS(each[0], bytes.NewBuffer([]byte{}))
		}
		if err != nil {
			return "", err
		}
	}

	hashes := strings.Join(hashArray, "\n")

	return hashes, nil
}

func catFile(fileHash string) (string, error) {
	var hash string
	var err error
	//CatFrom may have to contents here
	if logger.Level > 0 {
		hash, err = util.CatFromIPFS(fileHash, logger.Writer)
	} else {
		hash, err = util.CatFromIPFS(fileHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func listFile(objectHash string) (string, error) {
	var hash string
	var err error
	if logger.Level > 0 {
		hash, err = util.ListFromIPFS(objectHash, logger.Writer)
	} else {
		hash, err = util.ListFromIPFS(objectHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func listPinned() (string, error) {
	var hash string
	var err error
	if logger.Level > 0 {
		hash, err = util.ListPinnedFromIPFS(logger.Writer)
	} else {
		hash, err = util.ListPinnedFromIPFS(bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func rmAllPinned() (string, error) {

	hashList, err := listPinned()
	if err != nil {
		return "", err
	}

	hashArray := strings.Split(hashList, "\n")
	result := make([]string, len(hashArray))
	for i, hash := range hashArray {
		result[i], err = rmPinnedByHash(hash)
		if err != nil {
			return "", err
		}
	}
	hashes := strings.Join(result, "\n")
	return hashes, nil
}

func rmPinnedByHash(hash string) (string, error) {
	var err error
	if logger.Level > 0 {
		hash, err = util.RemovePinnedFromIPFS(hash, logger.Writer)
	} else {
		hash, err = util.RemovePinnedFromIPFS(hash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

//---------------------------------------------------------
// helpers

func writeCsv(hashArray, fileNames []string) error {

	strToWrite := make([][]string, len(hashArray))
	for i := range hashArray {
		strToWrite[i] = []string{hashArray[i], fileNames[i]}

	}

	csvfile, err := os.Create("ipfs_hashes.csv")
	if err != nil {
		return fmt.Errorf("error creating csv file:", err)
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	w.WriteAll(strToWrite)

	if err := w.Error(); err != nil {
		return fmt.Errorf("error writing csv: \n", err)
	}

	return nil
}
