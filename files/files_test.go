package files

import (
/*
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monax/cli/definitions"
	"github.com/monax/cli/log"
	"github.com/monax/cli/services"
	"github.com/monax/cli/testutil"
*/
)

/*
var (
	monaxDir     = filepath.Join(os.TempDir(), "monax")
	newDir       = filepath.Join(monaxDir, "addRecursively")
	fileInNewDir = filepath.Join(newDir, "recurse.toml")
	content      = "test contents"
	filename     = filepath.Join(monaxDir, "test-file.toml")
	port_to_use  = "8080"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	// Prevent CLI from starting IPFS.
	os.Setenv("MONAX_SKIP_ENSURE", "true")

	if port := os.Getenv("MONAX_CLI_TESTS_PORT"); port != "" {
		port_to_use = port
	}

	testutil.IfExit(testutil.Init(testutil.Pull{
		Services: []string{"ipfs"},
		Images:   []string{"ipfs"},
	}))
	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestPutFiles(t *testing.T) {
	testutil.FakeDefinitionFile(monaxDir, "test-file", content)

	do := definitions.NowDo()
	do.Name = filename
	do.IpfsPort = port_to_use

	hash := "QmcJdniiSKMp5az3fJvkbJTANd7bFtDoUkov3a8pkByWkv"

	// Fake IPFS server.
	os.Setenv("MONAX_IPFS_HOST", "http://127.0.0.1")
	ipfs := testutil.NewServer(fmt.Sprintf("127.0.0.1:%s", port_to_use))
	ipfs.SetResponse(testutil.ServerResponse{
		Code: http.StatusOK,
		Header: map[string][]string{
			"Ipfs-Hash": {hash},
		},
	})
	defer ipfs.Close()

	out, err := PutFiles(do)
	if err != nil {
		t.Fatalf("err putting files: %v\n", err)
	}

	if expected := "/ipfs/"; ipfs.Path() != expected {
		t.Fatalf("called the wrong endpoint; expected %v, got %v", expected, ipfs.Path())
	}

	if expected := "POST"; ipfs.Method() != expected {
		t.Fatalf("Used the wrong HTTP method; expected %v, got %v", expected, ipfs.Method())
	}

	if ipfs.Body() != content {
		t.Fatalf("Put the bad file; expected %q, got %q", content, ipfs.Body())
	}

	if hash != out {
		t.Fatalf("Hash mismatch; expected %q, got %q", hash, out)
	}
}

func TestGetFiles(t *testing.T) {
	var (
		hash     = "QmcJdniiSKMp5az3fJvkbJTANd7bFtDoUkov3a8pkByWkv"
		fileName = filepath.Join(monaxDir, "tset file.toml")
	)

	do := definitions.NowDo()
	do.Hash = hash
	do.Path = fileName
	do.IpfsPort = port_to_use

	// Fake IPFS server.
	os.Setenv("MONAX_IPFS_HOST", "http://127.0.0.1")
	ipfs := testutil.NewServer(fmt.Sprintf("127.0.0.1:%s", port_to_use))
	ipfs.SetResponse(testutil.ServerResponse{
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

	if returned := testutil.FileContents(fileName); content != returned {
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
	do.Hash = hash
	do.Path = monaxDir

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
}*/
