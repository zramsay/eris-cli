package util

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/log"
)

func GetStringResponse(question string, defalt string, reader *os.File) (string, error) {
	readr := bufio.NewReader(reader)
	log.Warn(question)

	text, _ := readr.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	if text == "" {
		return defalt, nil
	}
	return text, nil
}

func GetIntResponse(question string, defalt int, reader *os.File) (int, error) {
	readr := bufio.NewReader(reader)
	log.Warn(question)

	text, _ := readr.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	if text == "" {
		return defalt, nil
	}

	result, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, nil
	}

	return int(result), nil
}

// displays the question, scans for the response, if the response is an empty
// string will return default, otherwise will parseBool and return the result.
func GetBoolResponse(question string, defalt bool, reader *os.File) (bool, error) {
	var result bool
	readr := bufio.NewReader(reader)
	log.Warn(question)

	text, _ := readr.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	if text == "" {
		return defalt, nil
	}

	if text == "Yes" || text == "YES" || text == "Y" || text == "y" {
		result = true
	} else {
		result = false
	}

	return result, nil
}
