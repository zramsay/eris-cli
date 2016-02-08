package chains

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	def "github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteChainDefinitionFile(chainDef *def.Chain, fileName string) error {
	// writer := os.Stdout

	if filepath.Ext(fileName) == "" {
		fileName = chainDef.Name + ".toml"
		fileName = filepath.Join(ChainsPath, fileName)
	}

	writer, err := os.Create(fileName)
	defer writer.Close()
	if err != nil {
		return err
	}

	switch filepath.Ext(fileName) {
	case ".json":
		mar, err := json.MarshalIndent(chainDef, "", "  ")
		if err != nil {
			return err
		}
		mar = append(mar, '\n')
		writer.Write(mar)
	case ".yaml":
		mar, err := yaml.Marshal(chainDef)
		if err != nil {
			return err
		}
		mar = append(mar, '\n')
		writer.Write(mar)
	default:
		writer.Write([]byte("# This is a TOML config file.\n# For more information, see https://github.com/toml-lang/toml\n\n"))
		enc := toml.NewEncoder(writer)
		enc.Indent = ""
		writer.Write([]byte("name = \"" + chainDef.Name + "\"\n"))
		writer.Write([]byte("chain_id = \"" + chainDef.ChainID + "\"\n"))
		writer.Write([]byte("\n[service]\n"))
		enc.Encode(chainDef.Service)
		writer.Write([]byte("\n[maintainer]\n"))
		enc.Encode(chainDef.Maintainer)
	}
	return nil
}

func MakeGenesisFile(do *def.Do) error {

	//otherwise it'll start its own keys server that won't have the key needed...
	do.Name = "keys"
	IfExit(srv.EnsureRunning(do))

	doThr := def.NowDo()
	doThr.Chain.ChainType = "throwaway" //for teardown
	doThr.Name = "default"
	doThr.Chain.Name = "default" //for teardown
	doThr.Operations.ContainerNumber = 1
	doThr.Operations.PublishAllPorts = true

	log.WithField("=>", doThr.Name).Info("Making genesis.json file. Starting chain")
	if er := NewChain(doThr); er != nil {
		return fmt.Errorf("error starting chain %v\n", er)
	}

	doThr.Operations.Args = []string{"mintgen", "known", do.Chain.Name, fmt.Sprintf("--pub=%s", do.Pubkey)}

	// pipe this output to /chains/chainName/genesis.json
	_, err := ExecChain(doThr)
	if err != nil {
		log.Warnf("Executing chain error: %v", err)
		log.Warn("Cleaning up")
		doThr.Rm = true
		doThr.RmD = true
		if err := CleanUp(doThr); err != nil {
			return err
		}
	}

	doThr.Rm = true
	doThr.RmD = true
	return CleanUp(doThr) // doesn't clean up keys but that's ~ ok b/c it's about to be used...
}
