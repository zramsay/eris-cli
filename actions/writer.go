package actions

import (
	"os"
	"path/filepath"
	"strings"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteActionDefinitionFile(actDef *def.Action, fileName string) error {
	// writer := os.Stdout

	writer, err := os.Create(fileName)
	defer writer.Close()
	if err != nil {
		return err
	}

	if fileName == "" {
		fileName = strings.Replace(actDef.Name, " ", "_", -1) + ".toml"
	}

	switch filepath.Ext(fileName) {
	case ".toml":
		enc := toml.NewEncoder(writer)
		enc.Indent = ""
		writer.Write([]byte("name = \"" + actDef.Name + "\"\n"))
		writer.Write([]byte("services = [ \"" + strings.Join(actDef.Services, "\",\"") + "\" ]\n"))
		writer.Write([]byte("chains = [ \"" + strings.Join(actDef.Chains, "\",\"") + "\" ]\n"))
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
		writer.Write([]byte("\n[maintainer]\n"))
		enc.Encode(actDef.Maintainer)
		writer.Write([]byte("\n[location]\n"))
		enc.Encode(actDef.Location)
		writer.Write([]byte("\n[machine]\n"))
		enc.Encode(actDef.Machine)
	}
	return nil
}
