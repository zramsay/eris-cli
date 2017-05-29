package definitions

import "github.com/monax/monax/config"

// Compile request object
type Request struct {
	ScriptName      string                    `json:"name"`
	Language        string                    `json:"language"`
	Includes        map[string]*IncludedFiles `json:"includes"`  // our required files and metadata
	Libraries       string                    `json:"libraries"` // string of libName:LibAddr separated by comma
	Optimize        bool                      `json:"optimize"`  // run with optimize flag
	FileReplacement map[string]string         `json:"replacement"`
}

type BinaryRequest struct {
	BinaryFile string `json:"binary"`
	Libraries  string `json:"libraries"`
}

// this handles all of our imports
type IncludedFiles struct {
	ObjectNames []string `json:"objectNames"` //objects in the file
	Script      []byte   `json:"script"`      //actual code
}

const (
	SOLIDITY = "sol"
	SERPENT  = "se"
	LLL      = "lll"
)

type LangConfig struct {
	CacheDir     string   `json:"cache"`
	IncludeRegex string   `json:"regex"`
	CompileCmd   []string `json:"cmd"`
}

// Fill in the filename and return the command line args
func (l LangConfig) Cmd(includes []string, libraries string, optimize bool) (args []string) {
	for _, s := range l.CompileCmd {
		if s == "_" {
			if optimize {
				args = append(args, "--optimize")
			}
			if libraries != "" {
				args = append(args, "--libraries")
				args = append(args, libraries)
			}
			args = append(args, includes...)
		} else {
			args = append(args, s)
		}
	}
	return
}

// todo: add indexes for where to find certain parts in submatches (quotes, filenames, etc.)
// Global variable mapping languages to their configs
var Languages = map[string]LangConfig{
	LLL: {
		CacheDir:     config.LllcScratchPath,
		IncludeRegex: `\(include "(.+?)"\)`,
		CompileCmd: []string{
			"lllc",
			"_",
		},
	},
	SERPENT: {
		CacheDir:     config.SerpScratchPath,
		IncludeRegex: `create\(("|')(.+?)("|')\)`,
		CompileCmd: []string{
			"serpent",
			"mk_contract_info_decl",
			"_",
		},
	},
	SOLIDITY: {
		CacheDir:     config.SolcScratchPath,
		IncludeRegex: `import (.+?)??("|')(.+?)("|')(as)?(.+)?;`,
		CompileCmd: []string{
			"solc",
			"--combined-json", "bin,abi",
			"_",
		},
	},
}

// individual contract items
type SolcItem struct {
	Bin string `json:"bin"`
	Abi string `json:"abi"`
}

// full solc response object
type SolcResponse struct {
	Contracts map[string]*SolcItem `mapstructure:"contracts" json:"contracts"`
	Version   string               `mapstructure:"version" json:"version"` // json encoded
}

func BlankSolcItem() *SolcItem {
	return &SolcItem{}
}

func BlankSolcResponse() *SolcResponse {
	return &SolcResponse{
		Version:   "",
		Contracts: make(map[string]*SolcItem),
	}
}
