package initialize

func defKeys() string {
	return `[service]
name = "keys"

image = "eris/keys"
data_container = true
`
}

func defIpfs() string {
	return `name = "ipfs"

[service]
name = "ipfs"
image = "eris/ipfs"
data_container = true
ports = ["4001:4001", "5001:5001", "8080:8080"]
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

func defIpfs2() string {
	return `name = "ipfs"

  services = ["keys"]

[service]
name = "ipfs"
image = "eris/ipfs"
data_container = true
ports = ["4001:4001", "5001:5001", "8080:8080"]
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

func defAct() string {
	return `name = "do not use"
services = [ "ipfs" ]
chain = ""
steps = [
  "printenv",
  "echo hello",
  "echo goodbye"
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

func defChainConfig() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

moniker = "defaulttester.com"
seeds = ""
fast_sync = false
db_backend = "leveldb"
log_level = "debug"
node_laddr = ""
rpc_laddr = ""
`
}

func defChainGen() string {
	return `
{
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
}
`
}

func defChainKeys() string {
	return `
{
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
}
`
}

func defChainServConfig() string {
	return `
# This is a TOML config file.
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
