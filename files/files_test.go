package files

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/tests"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var erisDir string = filepath.Join(os.TempDir(), "eris")

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

	// Prevent CLI from starting IPFS.
	os.Setenv("ERIS_SKIP_ENSURE", "true")

	tests.IfExit(testsInit())
	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestPutFiles(t *testing.T) {
	var (
		content  = "test contents"
		filename = filepath.Join(erisDir, "test-file.toml")
	)
	tests.FakeDefinitionFile(erisDir, "test-file", content)

	do := definitions.NowDo()
	do.Name = filename
	log.WithField("=>", do.Name).Info("Putting file (from tests)")

	hash := "QmcJdniiSKMp5az3fJvkbJTANd7bFtDoUkov3a8pkByWkv"

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://127.0.0.1")
	ipfs := tests.NewServer("127.0.0.1:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Header: map[string][]string{
			"Ipfs-Hash": {hash},
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
	var (
		filename = filepath.Join(erisDir, "tset file.toml")
		content  = "test contents"
		hash     = "QmcJdniiSKMp5az3fJvkbJTANd7bFtDoUkov3a8pkByWkv"
	)
	do := definitions.NowDo()
	do.Name = hash
	do.Path = filename

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://127.0.0.1")
	ipfs := tests.NewServer("127.0.0.1:8080")
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

	if returned := tests.FileContents(filename); content != returned {
		fatal(t, fmt.Errorf("Returned unexpected content; expected %q, got %q", content, returned))
	}
}

func testsInit() error {
	if err := tests.TestsInit("files"); err != nil {
		return err
	}

	return nil
}
