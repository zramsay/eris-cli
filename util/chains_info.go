package util

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

// Maximum entries in the HEAD file
var MaxHead = 100

// check if given chain is known
func IsKnownChain(name string) bool {
	known := GetGlobalLevelConfigFilesByType("chains", false)
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}

// Change the head to null (no head)
func NullHead() error {
	return ChangeHead("")
}

// Get the current active chain (top of the HEAD file)
// Returns chain name
func GetHead() (string, error) {
	// TODO: only read the one line!
	f, err := ioutil.ReadFile(common.HEAD)
	if err != nil {
		return "", err
	}

	fspl := strings.Split(string(f), "\n")
	head := fspl[0]

	if head == "" {
		return "", fmt.Errorf("There is no chain checked out")
	}

	return head, nil
}

// Add a new entry (name) to the top of the HEAD file
// Expects the chain type and head (id) to be full (already resolved)
func ChangeHead(name string) error {
	if !IsKnownChain(name) && name != "" {
		logger.Debugf("Chain name not known. Not saving.\n")
		return nil
	}

	logger.Debugf("Chain name known (or blank). Saving to head file.\n")
	// read in the entire head file and clip
	// if we have reached the max length
	b, err := ioutil.ReadFile(common.HEAD)
	if err != nil {
		return err
	}
	bspl := strings.Split(string(b), "\n")
	var bsp string
	if len(bspl) >= MaxHead {
		bsp = strings.Join(bspl[:MaxHead-1], "\n")
	} else {
		bsp = string(b)
	}

	// add the new head
	var s string
	// handle empty head
	s = name + "\n" + bsp
	err = ioutil.WriteFile(common.HEAD, []byte(s), 0666)
	if err != nil {
		return err
	}

	logger.Debugf("Head file saved.\n")
	return nil
}
