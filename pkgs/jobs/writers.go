package jobs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const LogFileNameCSV = "jobs_output.csv"
const LogFileNameJSON = "jobs_output.json"

func ClearJobResults() error {
	if err := os.Remove(setJsonPath()); err != nil {
		return err
	}

	return os.Remove(setCsvPath())
}

// WriteJobResultCSV takes two strings and writes those to the delineated log
// file, which is currently epm.log in the same directory as the epm.yaml
func WriteJobResultCSV(name, result string) error {
	logFile := setCsvPath()

	var file *os.File
	var err error

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		file, err = os.Create(logFile)
	} else {
		file, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
	}

	if err != nil {
		return err
	}

	defer file.Close()

	text := fmt.Sprintf("%s,%s\n", name, result)
	if _, err = file.WriteString(text); err != nil {
		return err
	}

	return nil
}

func WriteJobResultJSON(results map[string]string) error {
	logFile := setJsonPath()

	file, err := os.Create(logFile)
	defer file.Close()

	res, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	if _, err = file.Write(res); err != nil {
		return err
	}

	return nil
}

func setJsonPath() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, LogFileNameJSON)
}

func setCsvPath() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, LogFileNameCSV)
}
