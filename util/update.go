package util

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func UpdateEris(branch string) {

	//check that git/go are installed
	CheckGitAndGo()

	//change pwd to eris/cli
	ChangeDirectory()

	if branch == "" {
		branch = "master"
	}

	CheckoutBranch(branch)
	PullBranch(branch)

	InstallEris()

	//fmt.Printf("Marmot update successful. Eris CLI Version is now: %s\n", VERSION)
}

func CheckGitAndGo() {
	stdOut1, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		fmt.Printf("ensure you have go installed:\n%s\n", string(stdOut1))
		os.Exit(1)
	}

	stdOut2, err := exec.Command("git", "version").CombinedOutput()
	if err != nil {
		fmt.Printf("ensure you have git installed:\n%v\n", string(stdOut2))
		os.Exit(1)
	}
}

func ChangeDirectory() {
	goPath := os.Getenv("GOPATH")
	dir := path.Join(goPath, "src/github.com/eris-ltd/eris-cli/")
	err := os.Chdir(dir)
	if err != nil {
		fmt.Printf("error changing directory\n%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("directory changed to:\n%s\n", dir)
}

func CheckoutBranch(branch string) {
	checkoutArgs := []string{"checkout", branch}

	stdOut, err := exec.Command("git", checkoutArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("error checking out %s:\n%s\n", branch, string(stdOut))
		os.Exit(1)
	}
	fmt.Printf("%s checked-out\n", branch)
}

func PullBranch(branch string) {
	pullArgs := []string{"pull", "origin", branch}

	stdOut, err := exec.Command("git", pullArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("error pulling from github:\n%s\n", string(stdOut))
		os.Exit(1)
	}
	fmt.Printf("%s pulled successfully\n", branch)
}

func InstallEris() {
	goArgs := []string{"install", "./cmd/eris"}
	stdOut, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("error with go install ./cmd/eris:\n%s\n", string(stdOut))
		os.Exit(1)
	}

}
