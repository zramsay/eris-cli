package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/eris-ltd/eris-pm/definitions"

	log "github.com/eris-ltd/eris-logger"
	// MARMOT: dependency on go-wire needs to be removed.
	"github.com/tendermint/go-wire"
)

const LogFileNameCSV = "epm.csv"
const LogFileNameJSON = "jobs_output.json"

// ------------------------------------------------------------------------
// Logging
// ------------------------------------------------------------------------

func ClearJobResults() error {
	if err := os.Remove(setJsonPath()); err != nil {
		return err
	}

	return os.Remove(setCsvPath())
}

func PrintPathPackage(do *definitions.Do) {
	log.WithField("=>", do.Compiler).Info("Using Compiler at")
	log.WithField("=>", do.Chain).Info("Using Chain at")
	log.WithField("=>", do.ChainID).Debug("With ChainID")
	log.WithField("=>", do.Signer).Info("Using Signer at")
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

// ------------------------------------------------------------------------
// Writers of Arbitrary stuff
// ------------------------------------------------------------------------

// FormatOutput formats arbitrary json in a viewable manner using reflection
func FormatOutput(args []string, i int, o interface{}) (string, error) {
	if len(args) < i+1 {
		return prettyPrint(o)
	}
	arg0 := args[i]
	v := reflect.ValueOf(o).Elem()
	name, err := fieldFromTag(v, arg0)
	if err != nil {
		return "", err
	}
	f := v.FieldByName(name)
	return prettyPrint(f.Interface())
}

func fieldFromTag(v reflect.Value, field string) (string, error) {
	iv := v.Interface()
	st := reflect.TypeOf(iv)
	for i := 0; i < v.NumField(); i++ {
		tag := st.Field(i).Tag.Get("json")
		if tag == field {
			return st.Field(i).Name, nil
		}
	}
	return "", fmt.Errorf("Invalid field name")
}

func prettyPrint(o interface{}) (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, wire.JSONBytes(o), "", "\t")
	if err != nil {
		return "", err
	}
	return string(prettyJSON.Bytes()), nil
}
