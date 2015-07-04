package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

func Initialize(toPull, verbose bool) error {
	if _, err := os.Stat(common.ErisRoot); err != nil {
		if err := common.InitErisDir(); err != nil {
			return fmt.Errorf("Could not Initialize the Eris Root Directory.\n%s\n", err)
		}
	} else {
		if verbose {
			fmt.Printf("Root eris directory (%s) already exists. Please type `eris` to see the help.\n", common.ErisRoot)
		}
	}

	if err := InitDefaultServices(toPull, verbose); err != nil {
		return fmt.Errorf("Could not instantiate default services.\n%s\n", err)
	}

	if verbose {
		fmt.Printf("Initialized eris root directory (%s) with default actions and service files.\n", common.ErisRoot)
	}

	// todo: when called from cli provide option to go on tour, like `ipfs tour`
	return nil
}

func InitDefaultServices(toPull, verbose bool) error {
	if toPull {
		if err := pullRepo("eris-services", common.ServicesPath, verbose); err != nil {
			if verbose {
				fmt.Println("Using default defs.")
			}
			if err2 := dropDefaults(); err2 != nil {
				return fmt.Errorf("Cannot pull: %s. %s.\n", err, err2)
			}
		} else {
			if err2 := pullRepo("eris-actions", common.ActionsPath, verbose); err2 != nil {
				return fmt.Errorf("Cannot pull actions: %s.\n", err2)
			}
		}
	} else {
		if err := dropDefaults(); err != nil {
			return err
		}
	}
	return nil
}

func pullRepo(name, location string, verbose bool) error {
	src := "https://github.com/eris-ltd/" + name
	c := exec.Command("git", "clone", src, location)
	if verbose {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func dropDefaults() error {
	if err := ipfsDef(); err != nil {
		return fmt.Errorf("Cannot add ipfs: %s.\n", err)
	}
	if err := edbDef(); err != nil {
		return fmt.Errorf("Cannot add erisdb: %s.\n", err)
	}
	if err := genDef(); err != nil {
		return fmt.Errorf("Cannot add default genesis: %s.\n", err)
	}
	if err := actDef(); err != nil {
		return fmt.Errorf("Cannot add default action: %s.\n", err)
	}
	return nil
}

func ipfsDef() error {
	if err := os.MkdirAll(common.ServicesPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(common.ServicesPath, "ipfs.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}
	ipfsD := defIpfs()
	writer.Write([]byte(ipfsD))
	return nil
}

func edbDef() error {
	if err := os.MkdirAll(common.ServicesPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(common.ServicesPath, "erisdb.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}
	edbD := defEdb()
	writer.Write([]byte(edbD))
	return nil
}

func genDef() error {
	genPath := filepath.Join(common.BlockchainsPath, "genesis")
	if err := os.MkdirAll(genPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(genPath, "default.json"))
	defer writer.Close()
	if err != nil {
		return err
	}
	gen := defGen()
	writer.Write([]byte(gen))
	return nil
}

func actDef() error {
	if err := os.MkdirAll(common.ActionsPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(common.ActionsPath, "do_not_use.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}
	act := defAct()
	writer.Write([]byte(act))
	return nil
}

func defIpfs() string {
	return `[service]
name = "ipfs"
image = "eris/ipfs"
data_container = true
ports = ["4001:4001", "5001", "8080:8080"]
user = "root"

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"

[location]
repository = "github.com/eris-ltd/eris-services"

[machine]
include = ["docker"]
requires = [""]
`
}

func defEdb() string {
	return `[service]
name           = "erisdb"
image          = "eris/erisdb:0.10"
ports          = [ "46656:46656", "46657:46657" ]
environment    = [ "TMROOT=/home/eris/.eris/blockchains/tendermint" ]
data_container = true

[manager]
fetch = "tendermint node --fast_sync" # we'd like this to stop when caught up!
start = "tendermint node"
new   = "tendermint node && last_pid=$! && sleep(1) && kill -KILL $last_pid"

[maintainer]
name  = "Eris Industries"
email = "support@erisindustries.com"

[location]
repository = "github.com/eris-ltd/eris-services"

[machine]
include  = [ "docker" ]
requires = [ "" ]
`
}

func defGen() string {
	return `{
  "genesis_time": "Wed Jun 24 21:54:02 +0000 2015",
  "chain_id": "etcb_testnet",
  "accounts": [
    {
      "address": "F3C0A608D9D942AF61A294CFF248F18A90A7A00A",
      "amount": 200000000
    },
    {
      "address": "BBFA7E58C4AB496FB4EEA429C76BE15678EEB189",
      "amount": 200000000
    },
    {
      "address": "F81CB9ED0A868BD961C4F5BBC0E39B763B89FCB6",
      "amount": 200000000
    },
    {
      "address": "1AC99AC0F9F321ADB73525DB2B4CB9D741BC7A97",
      "amount": 200000000
    },
    {
      "address": "BB8B65F1FC9EE8F90EE91EC74C80332D3EE589FA",
      "amount": 200000000
    },
    {
      "address": "964B1493BBE3312278B7DEB94C39149F7899A345",
      "amount": 200000000
    },
    {
      "address": "01101C8AA9C74021599B787729E018EC7AA1EB03",
      "amount": 200000000
    },
    {
      "address": "38277CF570DFA8EA77130EFBA6DD55C4E143C9C0",
      "amount": 200000000
    },
    {
      "address": "9F74F1ACCAE15B8D63E1BD0ED67C9D3A3EA71894",
      "amount": 200000000
    },
    {
      "address": "E0F8B081950D07C7FACE52E0AA740AC67A2D75EB",
      "amount": 200000000
    }
  ],
  "validators": [
    {
      "pub_key": [
        1,
        "F6C79CF0CB9D66B677988BCB9B8EADD9A091CD465A60542A8AB85476256DBA92"
      ],
      "amount": 1000000,
      "unbond_to": [
        {
          "address": "964B1493BBE3312278B7DEB94C39149F7899A345",
          "amount": 1000000
        }
      ]
    },
    {
      "pub_key": [
        1,
        "E15E88C226C5AEFF0597B4E71C9FEBF620538795C34CCCEB13D3CFECA8F6157B"
      ],
      "amount": 1000000,
      "unbond_to": [
        {
          "address": "01101C8AA9C74021599B787729E018EC7AA1EB03",
          "amount": 1000000
        }
      ]
    },
    {
      "pub_key": [
        1,
        "388180AC9AAF0C9A624DC0BC397A0FC0416E1713CD9181E41E8096DF5B6686FC"
      ],
      "amount": 1000000,
      "unbond_to": [
        {
          "address": "38277CF570DFA8EA77130EFBA6DD55C4E143C9C0",
          "amount": 1000000
        }
      ]
    },
    {
      "pub_key": [
        1,
        "A0B0501D148232AD06BF9C57361FE35B490E2920A7622F2840F88D4097B7ECB3"
      ],
      "amount": 1000000,
      "unbond_to": [
        {
          "address": "9F74F1ACCAE15B8D63E1BD0ED67C9D3A3EA71894",
          "amount": 1000000
        }
      ]
    },
    {
      "pub_key": [
        1,
        "6ED22473414B8DA547F5C781ADB12FB05BC6989A28BC0FADDA86D9831306F83C"
      ],
      "amount": 1000000,
      "unbond_to": [
        {
          "address": "E0F8B081950D07C7FACE52E0AA740AC67A2D75EB",
          "amount": 1000000
        }
      ]
    }
  ]
}
`
}

func defAct() string {
	return `name = "do not use"
services = [ "ipfs" ]
chains = [ "" ]
steps = [
  "printenv",
  "eris services export ipfs",
  "eris services -v import 1234 ipfs:$prev",
  "eris services known",
  "eris services ls",
  "eris services ps",
  "printenv"
]

[environment]
HELLO = "WORLD"

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"

[location]
repository = "github.com/eris-ltd/eris-cli"

[machine]
include = ["docker"]
requires = [""]
`
}
