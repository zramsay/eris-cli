package files

import (
	"bytes"
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
	logger.Debugf("Gonna Import a file =>\t\t%s:%v\n", do.Name, do.Path)
	err = importFile(do.Name, do.Path)
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
	logger.Debugf("Gonna Add a file =>\t\t%s:%v\n", do.Name, do.Path)
	hash, err = exportFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func PinFiles(do *definitions.Do) error {
	var hash string
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")
	logger.Debugf("Gonna Pin a file =>\t\t%s:%v\n", do.Name, do.Path)
	hash, err = pinFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
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
