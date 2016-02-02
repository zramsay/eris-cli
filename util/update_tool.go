package util

import (
	"fmt"
	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func downloadLatestRelease() (string, error) {
	latestURL := "https://github.com/eris-ltd/eris-cli/releases/latest"
	resp, err := http.Get(latestURL)
	if err != nil {
		log.Printf("could not retrieve latest eris release at %s", latestURL)
	}
	latestURL = resp.Request.URL.String()
	lastPos := strings.LastIndex(latestURL, "/")
	version := latestURL[lastPos+1:]
	platform := runtime.GOOS
	arch := runtime.GOARCH
	hostURL := "https://github.com/eris-ltd/eris-cli/releases/download/" + version + "/"
	filename := "eris_" + version[1:] + "_" + platform + "_" + arch
	fileURL := hostURL + filename
	switch platform {
	case "linux":
		filename += ".tar.gz"
	default:
		filename += ".zip"
	}

	ChangeDirectory("bin")
	var erisBin string
	output, err := os.Create(filename)
	// if we dont have permissions to create a file where eris cli exists, attempt to create file within HOME folder
	if err != nil {
		erisBin := filepath.Join(common.ScratchPath, "bin")
		if _, err = os.Stat(erisBin); os.IsNotExist(err) {
			err = os.MkdirAll(erisBin, 0755)
			if err != nil {
				log.Println("Error creating directory", erisBin, "Did not download binary. Exiting...")
				return "", err
			}
		}
		err = os.Chdir(erisBin)
		if err != nil {
			log.Println("Error changing directory to", erisBin, "Did not download binary. Exiting...")
			return "", err
		}
		output, err = os.Create(filename)
		if err != nil {
			log.Println("Error creating file", erisBin, "Exiting...")
			return "", err
		}
	}
	defer output.Close()
	fileResponse, err := http.Get(fileURL)
	if err != nil {
		log.Println("Error while downloading", filename, "-", err)
		return "", err
	}
	defer fileResponse.Body.Close()
	io.Copy(output, fileResponse.Body)
	if err != nil {
		log.Println("Error saving downloaded file", filename, "-", err)
		return "", err
	}
	erisLoc, _ := exec.LookPath("eris")
	if erisBin != "" {
		log.Println("downloaded eris binary", version, "for", platform, "to", erisBin, "\n Please manually move to", erisLoc)
	} else {
		log.Println("downloaded eris binary", version, "for", platform, "to", erisLoc)
	}
	var unzip string = "tar -xvf"
	if platform != "linux" {
		unzip = "unzip"
	}
	cmd := exec.Command("bin/sh", "-c", unzip, filename)
	err = cmd.Run()
	if err != nil {
		log.Println("unzipping", filename, "failed:", err)
	}
	return filename, nil
}

func UpdateEris(branch string, checkGit bool, checkGo bool) {

	//check that git/go are installed
	hasGit, hasGo := CheckGitAndGo(checkGit, checkGo)
	if hasGo == false {
		log.Println("Go is not installed. Downloading eris-cli binary...")
		_, err := downloadLatestRelease()
		if err != nil {
			log.Println("Latest binary failed to download with error:", err)
			log.Println("Exiting...")
			os.Exit(1)
		}
	} else if hasGit == false {
		log.Println("Git is not installed. Please install git before continuing.")
		log.Println("Exiting...")
		os.Exit(1)
	}

	//checks for deprecated dir names and renames them
	err := MigrateDeprecatedDirs(common.DirsToMigrate, false) // false = no prompt
	if err != nil {
		log.Warnf("Directory migration error: %v", err)
		log.Warn("Continuing with update without migration")
	}

	//change pwd to eris/cli
	ChangeDirectory("src")

	if branch == "" {
		branch = "master"
	}

	CheckoutBranch(branch)
	PullBranch(branch)

	InstallEris()
	ver := version() //because version.Version will be in RAM.

	log.WithField("=>", ver).Warn("The marmots have updated Eris successfully")
}

func CheckGitAndGo(git, gO bool) (bool, bool) {
	hasGit := false
	hasGo := false
	if git {
		stdOut1, err := exec.Command("git", "version").CombinedOutput()
		if err != nil {
			log.WithField("version", string(stdOut1)).Warn("Ensure you have git installed.")
		} else {
			hasGit = true
		}
	}
	if gO {
		stdOut2, err := exec.Command("go", "version").CombinedOutput()
		if err != nil {
			log.WithField("version", string(stdOut2)).Warn("Ensure you have Go installed.")
		} else {
			hasGo = true
		}
	}
	return hasGit, hasGo
}

func ChangeDirectory(to string) {
	if to == "bin" {
		erisLoc, err := exec.LookPath("eris")
		if err != nil {
			log.Fatalf("Error finding eris binary: %v", err)
		}
		err = os.Chdir(filepath.Dir(erisLoc))
		if err != nil {
			log.Fatalf("Error changing directory: %v", err)
		}
		log.WithField("dir", erisLoc).Debug("Directory changed to")
	} else if to == "src" {
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			log.Fatal("You do not have $GOPATH set. Please make sure this is set and rerun the command.")
		}

		dir := filepath.Join(goPath, "src", "github.com", "eris-ltd", "eris-cli")
		err := os.Chdir(dir)

		if err != nil {
			log.Fatalf("Error changing directory: %v")
		}
		log.WithField("dir", dir).Debug("Directory changed to")
	}
}

func CheckoutBranch(branch string) {
	checkoutArgs := []string{"checkout", branch}

	stdOut, err := exec.Command("git", checkoutArgs...).CombinedOutput()
	if err != nil {
		log.WithField("branch", branch).Fatalf("Error checking out branch: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch checked-out")
}

func PullBranch(branch string) {
	pullArgs := []string{"pull", "origin", branch}

	stdOut, err := exec.Command("git", pullArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error pulling from GitHub: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch pulled successfully")
}

func InstallEris() {
	goArgs := []string{"install", "./cmd/eris"}

	stdOut, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error with go install ./cmd/eris: %v", string(stdOut))
	}

	log.Debug("Go install worked correctly")
}

func version() string {
	verArgs := []string{"version"}

	stdOut, err := exec.Command("eris", verArgs...).CombinedOutput()
	if err != nil {
		common.IfExit(fmt.Errorf("error getting version:\n%s\n", string(stdOut)))
	}
	return string(stdOut)

}
