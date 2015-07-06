package util

import (
	def "github.com/eris-ltd/eris-cli/definitions"
)

// need to be alot smarter with this
func OverwriteOps(opsBase, opsOver *def.Operation) {
	if opsOver.PublishAllPorts {
		opsBase.PublishAllPorts = true
	}
}
