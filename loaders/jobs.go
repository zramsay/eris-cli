package loaders

import (
	"fmt"
	"path/filepath"

	"github.com/monax/cli/definitions"
	"github.com/monax/cli/log"
	"github.com/monax/cli/pkgs/jobs"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/loggers"

	"github.com/spf13/viper"
)

func LoadJobs(do *definitions.Do) (*jobs.Jobs, error) {
	log.Info("Loading Job Runner File...")
	var fileName = do.YAMLPath
	var jobset = jobs.EmptyJobs()

	burrowClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	_, chainID, _, err := burrowClient.ChainId()
	log.WithField("=>", chainID).Info("Loaded chain ID")

	jobset.NodeClient = burrowClient
	jobset.KeyClient = keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	jobset.PublicKey = do.PublicKey
	jobset.DefaultAddr = do.DefaultAddr
	jobset.OutputFormat = do.DefaultOutput
	jobset.DefaultSets = do.DefaultSets
	jobset.Overwrite = do.Overwrite
	jobset.DefaultAmount = do.DefaultAmount
	jobset.DefaultFee = do.DefaultFee
	jobset.DefaultGas = do.DefaultGas
	jobset.JobMap = make(map[string]*jobs.JobResults)
	jobset.AbiMap = make(map[string]string)
	jobset.AbiPath = do.ABIPath
	jobset.BinPath = do.BinPath

	jobset.ChainID = chainID

	if err != nil {
		return nil, err
	}

	var loadJobs = viper.New()

	// setup file
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("Sorry, the marmots were unable to find the absolute path to the monax jobs file.")
	}

	path := filepath.Dir(abs)
	file := filepath.Base(abs)
	extName := filepath.Ext(file)
	bName := file[:len(file)-len(extName)]
	log.WithFields(log.Fields{
		"path": path,
		"name": bName,
	}).Debug("Loading jobs file")

	loadJobs.AddConfigPath(path)
	loadJobs.SetConfigName(bName)

	loadJobs.AddConfigPath(path)
	loadJobs.SetConfigName(bName)

	// load file
	if err := loadJobs.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Sorry, the marmots were unable to load the job runner file. Please check your path: %v", err)
	}

	// marshall file
	if err := loadJobs.Unmarshal(jobset); err != nil {
		return nil, fmt.Errorf(`Sorry, the marmots could not figure that job runner file out. 
			Please check that your jobs config is properly formatted: %v`, err)
	}

	return jobset, nil
}
