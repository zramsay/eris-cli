package initialize

import ()

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

func defEdb() string {
	// TODO: remove this. we should be hard coding these defaults...
	return `[service]
name           = "erisdb"
image          = "eris/erisdb:develop"
ports          = [ "46656", "46657" ]
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
