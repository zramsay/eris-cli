package util

import (
	"fmt"
	"regexp"

	"github.com/eris-ltd/eris-pm/definitions"
)

func BundleHttpPathCorrect(do *definitions.Do) {
	do.Chain = HttpPathCorrect(do.Chain, "tcp", true)
	do.Signer = HttpPathCorrect(do.Signer, "http", false)
	do.Compiler = HttpPathCorrect(do.Compiler, "http", false)
}

func HttpPathCorrect(oldPath, requiredPrefix string, trailingSlash bool) string {
	var newPath string
	protoReg := regexp.MustCompile(fmt.Sprintf("%ss*://.*", requiredPrefix))
	trailer := regexp.MustCompile("/$")

	if !protoReg.MatchString(oldPath) {
		newPath = requiredPrefix + "://" + oldPath
	} else {
		newPath = oldPath
	}

	if trailingSlash {
		if !trailer.MatchString(newPath) {
			newPath = newPath + "/"
		}
	}

	return newPath
}
