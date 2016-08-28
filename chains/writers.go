package chains

import (
	//"encoding/json"
	//"os"
	//"path/filepath"
	"fmt"
	"io/ioutil"

	//def "github.com/eris-ltd/eris-cli/definitions"

	//. "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
	//"github.com/BurntSushi/toml"
	//"gopkg.in/yaml.v2"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteChainDefinitionFile(pathToConfig, fileName string) error {

	// TODO refactor so that it just writes a file that can tell cli which
	// dir to look for the config file in
	if err := ioutil.WriteFile(fileName, []byte(pathToConfig), 0666); err != nil {
		return err
	}
	log.Warn(fmt.Sprintf("%s written with path to config.toml:\n%s", fileName, pathToConfig))
	return nil

	// writer := os.Stdout

	/*if filepath.Ext(fileName) == "" {
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
	return nil*/
}
