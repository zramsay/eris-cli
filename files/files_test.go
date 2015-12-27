package files

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/logger"

	tests "github.com/eris-ltd/eris-cli/testutils"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var file string
var content string = "test content\n"

var DEAD bool // XXX: don't double panic (TODO: Flushing twice blocks)
func fatal(t *testing.T, err error) {
	if !DEAD {
		tests.TestsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		erisDir = os.Getenv("HOME")
	}

        // Prevent CLI from starting IPFS.
        os.Setenv("ERIS_SKIP_ENSURE", "true")

	file = path.Join(erisDir, "temp")

	tests.IfExit(testsInit())
	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestPutFiles(t *testing.T) {
	do := definitions.NowDo()
	do.Name = file
	log.WithField("=>", do.Name).Info("Putting file (from tests)")

	hash := "QmcJdniiSKMp5az3fJvkbJTANd7bFtDoUkov3a8pkByWkv"

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://localhost")
	ipfs := tests.NewServer("localhost:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Header: map[string][]string{
			"Ipfs-Hash": []string{hash},
		},
	})
	defer ipfs.Close()

	if err := PutFiles(do); err != nil {
		fatal(t, err)
	}

	if expected := "/ipfs/"; ipfs.Path() != expected {
		fatal(t, fmt.Errorf("Called the wrong endpoint; expected %v, got %v\n", expected, ipfs.Path()))
	}

	if expected := "POST"; ipfs.Method() != expected {
		fatal(t, fmt.Errorf("Used the wrong HTTP method; expected %v, got %v\n", expected, ipfs.Method()))
	}

	if ipfs.Body() != content {
		fatal(t, fmt.Errorf("Put the bad file; expected %q, got %q\n", content, ipfs.Body()))
	}

	if hash != do.Result {
		fatal(t, fmt.Errorf("Hash mismatch; expected %q, got %q\n", hash, do.Result))
	}

	log.WithField("result", do.Result).Debug("Finished putting a file")
}

func TestGetFiles(t *testing.T) {
	fileName := strings.Replace(file, "temp", "pmet", 1)
	hash := "QmcJdniiSKMp5az3fJvkbJTANd7bFtDoUkov3a8pkByWkv"
	do := definitions.NowDo()
	do.Name = hash
	do.Path = fileName

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://localhost")
	ipfs := tests.NewServer("localhost:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Body: content,
	})
	defer ipfs.Close()

	if err := GetFiles(do); err != nil {
		fatal(t, err)
	}

	if expected := "/ipfs/" + hash; ipfs.Path() != expected {
		fatal(t, fmt.Errorf("Called the wrong endpoint; expected %v, got %v\n", expected, ipfs.Path()))
	}

	if expected := "GET"; ipfs.Method() != expected {
		fatal(t, fmt.Errorf("Used the wrong HTTP method; expected %v, got %v\n", expected, ipfs.Method()))
	}

	f, err := os.Open(fileName)
	if err != nil {
		fatal(t, err)
	}

	contentPuted, err := ioutil.ReadAll(f)
	if err != nil {
		fatal(t, err)
	}

	if string(contentPuted) != content {
		fatal(t, fmt.Errorf("Returned unexpected content; expected: %q, got %q", content, string(contentPuted)))
	}
}

func testsInit() error {
	if err := tests.TestsInit("files"); err != nil {
		return err
	}

	f, err := os.Create(file)
	tests.IfExit(err)
	f.Write([]byte(content))

	return nil
}
