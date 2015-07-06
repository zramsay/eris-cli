package files

import (
	"bytes"

	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

func GetFilesRaw(hash, fileName string) error {
	ipfsService, err := services.LoadServiceDefinition("ipfs", 1)
	if err != nil {
		return err
	}

	if services.IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
		logger.Infoln("IPFS is running. Adding now.")

		err := importFile(hash, fileName)
		if err != nil {
			return err
		}

	} else {
		logger.Infoln("IPFS is not running. Starting now.")
		err := services.StartServiceByService(ipfsService.Service, ipfsService.Operations, []string{})
		if err != nil {
			return err
		}

		err = importFile(hash, fileName)
		if err != nil {
			return err
		}
	}
	return nil
}

func PutFilesRaw(fileName string) (string, error) {
	var hash string
	var err error

	ipfsService, err := services.LoadServiceDefinition("ipfs", 1)
	if err != nil {
		return "", err
	}

	if services.IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
		logger.Infoln("IPFS is running. Adding now.")

		hash, err = exportFile(fileName)
		if err != nil {
			return "", err
		}

	} else {
		logger.Infoln("IPFS is not running. Starting now.")
		if err := services.StartServiceByService(ipfsService.Service, ipfsService.Operations, []string{}); err != nil {
			return "", err
		}

		hash, err = exportFile(fileName)
		if err != nil {
			return "", err
		}
	}
	return hash, nil
}

func importFile(hash, fileName string) error {
	var err error
	if logger.Level > 0 {
		err = util.GetFromIPFS(hash, fileName, logger.Writer)
	} else {
		err = util.GetFromIPFS(hash, fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return err
	}
	return nil
}

func exportFile(fileName string) (string, error) {
	var hash string
	var err error

	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}

	return hash, nil
}
