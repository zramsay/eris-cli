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

	if do.AddDir {
		logger.Debugf("Adding directory recursively")
		hashes, err := exportDir(do.Name)
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
	//XXX need logic to prevent do.AddDir & do.Gateway from conflict
	if do.Gateway {
		logger.Debugf("Also pinning it to gateway (ipfs.erisbootstrap.sexy) =>\t\t%s:%v\n", do.Name, do.Path)
	}

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
	//could also take flag instead of false
	hash, err = pinFile(do.Name, false)
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

func ListPinned(do *definitions.Do) error {
	var hash string
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		return err
	}
	logger.Infoln("IPFS is running.")
	logger.Debugf("Listing files pinned locally")
	hash, err = listPinned()
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

func exportDir(dirName string) (string, error) {
	var hashes string
	var err error

	if logger.Level > 0 {
		hashes, err = util.SendDirToIPFS(dirName, logger.Writer)
	} else {
		hashes, err = util.SendDirToIPFS(dirName, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	//hashes := strings.Split whatever returns from the shell
	//do.Result = hashes

	return hashes, nil
}

func pinFile(fileHash string, gateway bool) (string, error) {
	var hash string
	var err error

	if logger.Level > 0 {
		hash, err = util.PinToIPFS(fileHash, gateway, logger.Writer)
	} else {
		hash, err = util.PinToIPFS(fileHash, gateway, bytes.NewBuffer([]byte{}))
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
