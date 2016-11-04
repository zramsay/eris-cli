package util

import (
	"fmt"
	"strings"
)

const (
	No = iota
	Yes
)

func QueryYesOrNo(question string, defaults ...int) int {
	defaultValue := No
	if len(defaults) > 0 {
		defaultValue = defaults[0]

		if defaultValue == Yes {
			question += fmt.Sprintf(" (default: yes) ")
		} else {
			question += fmt.Sprintf(" (default: no) ")
		}
	}

	fmt.Printf("%s (y/n): ", question)
	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return defaultValue
	}
	input = strings.ToLower(input)
	if input[0] == 'y' {
		return Yes
	}
	if input[0] == 'n' {
		return No
	}
	return defaultValue
}
