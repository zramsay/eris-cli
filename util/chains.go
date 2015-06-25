package util

import (
	"fmt"
	"strings"
)

// TODO: bring in all those nice epm chains functions
// for handling chain names/types/ids/paths

func SplitChainTypeID(chain string) (string, string, error) {
	spl := strings.Split(chain, "/")
	if len(spl) != 2 {
		return "", "", fmt.Errorf("Invalid chain name (%s), expected <chainType>:<chainID>", chain)
	}
	return spl[0], spl[1], nil
}
