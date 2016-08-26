package initialize

import (
	"fmt"
)

//TODO replace these with `make` an use that for tests...?
func DefChainGen() string {
	return `
{
  "chain_id": "my_tests",
  "accounts": [
    {
      "address": "0000000000000000000000000000000000000001",
      "amount": 690000000000
    },
    {
      "address": "0000000000000000000000000000000000000002",
      "amount": 565000000000
    },
    {
      "address": "0000000000000000000000000000000000000003",
      "amount": 525000000000
    },
    {
      "address": "0000000000000000000000000000000000000004",
      "amount": 110000000000
    },
    {
      "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
      "amount": 999999999999
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
          "address": "37236DF251AB70022B1DA351F08A20FB52443E37",
          "amount": 5000000000
        }
      ]
    }
  ]
}
`
}

// different from genesis above! -- used for testing
//[zr] leave for testing - for now ?
var DefaultPubKeys = []string{"CB3688B7561D488A2A4834E1AEE9398BEF94844D8BDBBCA980C11E3654A45906"}

func DefChainCSV() string {
	return fmt.Sprintf("%s,", DefaultPubKeys[0])
}

//use tool to gen one of these...
func DefChainKeys() string {
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
