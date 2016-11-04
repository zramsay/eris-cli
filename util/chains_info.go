package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/log"
)

// Maximum entries in the HEAD file
var MaxHead = 100

// Change the head to null (no head)
func NullHead() error {
	return ChangeHead("")
}

// Get the current active chain (top of the HEAD file)
// Returns chain name
func GetHead() (string, error) {
	// TODO: only read the one line!
	f, err := ioutil.ReadFile(config.HEAD)
	if os.IsNotExist(err) {
		if _, err := os.Create(config.HEAD); err != nil {
			return "", err
		}
	} else {
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
	if !IsChain(name, false) && name != "" {
		log.Debug("Chain name not known. Not saving")
		return nil
	}

	log.Debug("Chain name known (or blank). Saving to head file")
	// read in the entire head file and clip
	// if we have reached the max length
	b, err := ioutil.ReadFile(config.HEAD)
	if os.IsNotExist(err) {
		if _, err := os.Create(config.HEAD); err != nil {
			return err
		}
	} else if err != nil {
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
	if err := ioutil.WriteFile(config.HEAD, []byte(s), 0666); err != nil {
		return err
	}

	log.Debug("Head file saved")
	return nil
}
