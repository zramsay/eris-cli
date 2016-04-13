package files

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"

	log "github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/common/go/log"
)

var (
	erisDir      = filepath.Join(os.TempDir(), "eris")
	newDir       = filepath.Join(erisDir, "addRecursively")
	fileInNewDir = filepath.Join(newDir, "recurse.toml")
	content      = "test contents"
	filename     = filepath.Join(erisDir, "test-file.toml")
)

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	// Prevent CLI from starting IPFS.
	os.Setenv("ERIS_SKIP_ENSURE", "true")

	tests.IfExit(tests.TestsInit("files"))
	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestPutFiles(t *testing.T) {
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
		t.Fatalf("err putting files: %v\n", err)
	}

	if expected := "/ipfs/"; ipfs.Path() != expected {
		t.Fatalf("called the wrong endpoint; expected %v, got %v\n", expected, ipfs.Path())
	}

	if expected := "POST"; ipfs.Method() != expected {
		t.Fatalf("Used the wrong HTTP method; expected %v, got %v\n", expected, ipfs.Method())
	}

	if ipfs.Body() != content {
		t.Fatalf("Put the bad file; expected %q, got %q\n", content, ipfs.Body())
	}

	if hash != do.Result {
		t.Fatalf("Hash mismatch; expected %q, got %q\n", hash, do.Result)
	}

	log.WithField("result", do.Result).Debug("Finished putting a file")
}

func TestGetFiles(t *testing.T) {
	var (
		hash     = "QmcJdniiSKMp5az3fJvkbJTANd7bFtDoUkov3a8pkByWkv"
		fileName = filepath.Join(erisDir, "tset file.toml")
	)

	do := definitions.NowDo()
	do.Name = hash
	do.Path = fileName

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://127.0.0.1")
	ipfs := tests.NewServer("127.0.0.1:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Body: content,
	})
	defer ipfs.Close()

	passed := false
	for i := 0; i < 5; i++ {
		if err := GetFiles(do); err != nil {
			time.Sleep(2 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}
	if !passed {
	// final time will throw
		if err := GetFiles(do); err != nil {
			t.Fatalf("err getting files %v\n", err)
		}
	}

	if expected := "/ipfs/" + hash; ipfs.Path() != expected {
		t.Fatalf("Called the wrong endpoint; expected %v, got %v\n", expected, ipfs.Path())
	}

	if expected := "GET"; ipfs.Method() != expected {
		t.Fatalf("Used the wrong HTTP method; expected %v, got %v\n", expected, ipfs.Method())
	}

	if returned := tests.FileContents(fileName); content != returned {
		t.Fatalf("Returned unexpected content; expected %q, got %q", content, returned)
	}
}

func TestPutAndGetDirectory(t *testing.T) {
	buf, err := services.ExecHandler("ipfs", []string{"bash", "-c", "ipfs init; ipfs get `ipfs add -qr /root`"})
	if err != nil {
		t.Fatalf("Failed to put and get a directory from IPFS: %v : %s", err, buf.String())
	}
	fmt.Println(buf.String())
}

func testGetDirectoryFromIPFS(t *testing.T) {
	var err error
	hash := "QmYwjCPtWkduz81UnAqMJYCag5pock5y2S8yZQEd4qoyzf"

	do := definitions.NowDo()
	do.Name = hash
	do.Path = erisDir

	passed := false
	for i := 0; i < 10; i++ { //usually needs 3-4
		_, err = importDirectory(do)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}

	if !passed {
		_, err = importDirectory(do)
		if err != nil {
			t.Fatalf("error putting dir to IPFS: %v\n", err)

		}
	}
}

// get a dir up in there
// adapted from agent/agent_test.go
// eventually deduplicate
func testPutDirectoryToIPFS(t *testing.T) {
	var err error

	if err := os.MkdirAll(newDir, 0777); err != nil {
		t.Fatalf("err mkdir: %v\n", err)
	}

	if err := ioutil.WriteFile(fileInNewDir, []byte(content), 0777); err != nil {
		t.Fatalf("err writing file: %v\n", err)
	}

	do := definitions.NowDo()
	do.Name = newDir

	passed := false
	for i := 0; i < 10; i++ { //usually needs 3-4
		_, err = exportDirectory(do)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}

	if !passed {
		_, err = exportDirectory(do)
		if err != nil {
			t.Fatalf("error putting dir to IPFS: %v\n", err)

		}
	}
}

func testStartIPFS(t *testing.T) {
	do := definitions.NowDo()
	do.Name = "ipfs"
	do.Operations.Args = []string{"ipfs"}
	do.Operations.PublishAllPorts = false
	if err := services.StartService(do); err != nil {
		t.Fatalf("expected service to start, got %v", err)
	}
	// because it might help ... ?
	time.Sleep(1 * time.Second)
}

func testKillIPFS(t *testing.T) {
	do := definitions.NowDo()
	do.Name = "ipfs"
	do.Operations.Args = []string{"ipfs"}
	do.Rm = true
	do.RmD = true
	if err := services.KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}
}
