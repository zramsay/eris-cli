package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteActionDefinitionFile(actDef *def.Action, fileName string) error {
	if strings.Contains(fileName, " ") {
		fileName = strings.Replace(actDef.Name, " ", "_", -1)
	}
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	log.WithFields(log.Fields{
		"action": actDef.Name,
		"file":   fileName,
	}).Debug("Writing action definition file")

	writer, err := os.Create(fileName)
	defer writer.Close()
	if err != nil {
		return err
	}

	switch filepath.Ext(fileName) {
	case ".json":
		mar, err := json.MarshalIndent(actDef, "", "  ")
		if err != nil {
			return err
		}
		mar = append(mar, '\n')
		writer.Write(mar)
	case ".yaml":
		mar, err := yaml.Marshal(actDef)
		if err != nil {
			return err
		}
		mar = append(mar, '\n')
		writer.Write(mar)
	default:
		writer.Write([]byte("# This is a TOML config file.\n# For more information, see https://github.com/toml-lang/toml\n\n"))
		enc := toml.NewEncoder(writer)
		enc.Indent = ""
		writer.Write([]byte("name = \"" + actDef.Name + "\"\n"))
		writer.Write([]byte("chain = \"" + actDef.Chain + "\"\n"))
		writer.Write([]byte("steps = [ \n"))
		for _, s := range actDef.Steps {
			if strings.Contains(s, "\"") {
				s = strings.Replace(s, "\"", "\\\"", -1)
			}
			writer.Write([]byte("  \"" + s + "\",\n"))
		}
		writer.Write([]byte("] \n"))
		writer.Write([]byte("\n[environment]\n"))
		enc.Encode(actDef.Environment)
		writer.Write([]byte("[dependencies]\n"))
		if actDef.Dependencies != nil {
			if len(actDef.Dependencies.Services) != 0 || len(actDef.Dependencies.Chains) != 0 {
				enc.Encode(actDef.Dependencies)
			}
		}
		writer.Write([]byte("\n[maintainer]\n"))
		enc.Encode(actDef.Maintainer)
		writer.Write([]byte("\n[location]\n"))
		enc.Encode(actDef.Location)
		writer.Write([]byte("\n[machine]\n"))
		enc.Encode(actDef.Machine)
	}
	return nil
}
