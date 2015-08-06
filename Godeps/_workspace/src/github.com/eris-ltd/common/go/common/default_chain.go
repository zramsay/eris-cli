package common

import (
	"fmt"
	"path/filepath"
)

func InitDefaultChains(path string, verbose bool) error {
	if verbose {
		fmt.Println("Installing default chain defs.")
	}
	if err := dropChainDefaults(); err != nil {
		return fmt.Errorf("Cannot write default chains: %s.\n", err)
	}
	return nil
}

func dropChainDefaults() error {
	if err := WriteFile(produceDefaultConfig(), filepath.Join(ChainsConfigPath, "default", "config.toml")); err != nil {
		return err
	}
	if err := WriteFile(produceDefaultGenesis(), filepath.Join(ChainsConfigPath, "default", "genesis.json")); err != nil {
		return err
	}
	if err := WriteFile(produceDefaultPrivVal(), filepath.Join(ChainsConfigPath, "default", "priv_validator.json")); err != nil {
		return err
	}
	if err := WriteFile(produceServerConf(), filepath.Join(ChainsConfigPath, "default", "server_conf.toml")); err != nil {
		return err
	}
	return nil
}

func produceDefaultConfig() string {
	return `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

moniker = "anothertester"
seeds = ""
fast_sync = false
db_backend = "leveldb"
log_level = "debug"
node_laddr = ""`
}

func produceDefaultGenesis() string {
	return `{
  "chain_id": "my_tests",
  "accounts": [
    {
      "address": "F81CB9ED0A868BD961C4F5BBC0E39B763B89FCB6",
      "amount": 690000000000
    },
    {
      "address": "0000000000000000000000000000000000000002",
      "amount": 565000000000
    },
    {
      "address": "9E54C9ECA9A3FD5D4496696818DA17A9E17F69DA",
      "amount": 525000000000
    },
    {
      "address": "0000000000000000000000000000000000000004",
      "amount": 110000000000
    },
    {
      "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
      "amount": 110000000000
    }
  ],
  "validators": [
    {
      "pub_key": [
        1,
        "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
      ],
      "amount": 5000000000,
      "unbond_to": [
        {
          "address": "93E243AC8A01F723DE353A4FA1ED911529CCB6E5",
          "amount": 5000000000
        }
      ]
    }
  ]
}`
}

func produceDefaultPrivVal() string {
	return `{
  "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
  "pub_key": [
    1,
    "CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
  ],
  "priv_key": [
    1,
    "6B72D45EB65F619F11CE580C8CAED9E0BADC774E9C9C334687A65DCBAD2C4151CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"
  ],
  "last_height": 0,
  "last_round": 0,
  "last_step": 0
}`
}

func produceServerConf() string {
	return `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

[bind]
address=""
port=1337

[TLS]
tls=false
cert_path=""
key_path=""

[CORS]
enable=false
allow_origins=[]
allow_credentials=false
allow_methods=[]
allow_headers=[]
expose_headers=[]
max_age=0

[HTTP]
json_rpc_endpoint="/rpc"

[web_socket]
websocket_endpoint="/socketrpc"
max_websocket_sessions=50
read_buffer_size=2048
write_buffer_size=2048

[logging]
console_log_level="info"
file_log_level="warn"
log_file=""
`
}
