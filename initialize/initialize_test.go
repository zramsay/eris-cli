package initialize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris/config"
	"github.com/eris-ltd/eris/log"
	"github.com/eris-ltd/eris/util"
	"github.com/eris-ltd/eris/version"
)

const serviceToNeverUserToml = `

`

func TestGetServiceDefinitionFileBytes(t *testing.T) {

	const name = "do_not_use_ever"

	// send to
	testBytes, err := getServiceDefinitionFileBytes(name)
	if err != nil {
		t.Fatalf(err)
	}

	// compare testBytes to serviceToNeverUseToml

}
