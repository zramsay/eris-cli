package chains

import (
	"encoding/json"
	"os"
	"path/filepath"

	def "github.com/eris-ltd/eris-cli/definitions"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteChainDefinitionFile(chainDef *def.Chain, fileName string) error {
	// writer := os.Stdout

	if filepath.Ext(fileName) == "" {
		fileName = chainDef.Name + ".toml"
		fileName = filepath.Join(BlockchainsPath, fileName)
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
		writer.Write([]byte("type = \"" + chainDef.Type + "\"\n"))
		writer.Write([]byte("chain_id = \"" + chainDef.ChainID + "\"\n"))
		writer.Write([]byte("\n[service]\n"))
		enc.Encode(chainDef.Service)
		writer.Write([]byte("\n[manager]\n"))
		enc.Encode(chainDef.Manager)
	}
	return nil
}
