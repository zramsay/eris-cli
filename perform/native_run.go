package perform

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

// Helper function to execute the cli commands and send back
// the result of the command to the caller.
func NativeCommand(name string, args ...string) {
	product, err := NativeCommandRaw(name, args...)
	if err != nil {
		// TODO: fix
		return
	}
	fmt.Print(product)
}

// Assembles and Executes the command.
func NativeCommandRaw(name string, args ...string) (string, error) {
	var cmd *exec.Cmd

	cmd = exec.Command(name, args...)

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		// TODO: handle errors
		log.Fatal(err)
		return "", err
	}

	if errOut.String() != "" {
		err := fmt.Errorf(errOut.String())
		// TODO: handle
		log.Fatal(err)
		return "", err
	}

	return out.String(), nil
}
