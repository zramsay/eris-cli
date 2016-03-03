package util

import (
	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
)

// Labels returns a map with container labels, based on the container
// short name and ops settings.
//
//  ops.SrvContainerName  - container name
//  ops.ContainerType     - container type
//
func Labels(name string, ops *def.Operation) map[string]string {
	labels := ops.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[def.Namespace+":"+def.LabelEris] = "true"
	labels[def.Namespace+":"+def.LabelShortName] = name
	labels[def.Namespace+":"+def.LabelType] = ops.ContainerType
	labels[def.Namespace+":"+def.LabelNumber] = "1"

	if user, _, err := config.GitConfigUser(); err == nil {
		labels[def.Namespace+":"+def.LabelUser] = user
	}

	return labels
}

// SetLabel returns a labels map with additional label name and value.
func SetLabel(labels map[string]string, name, value string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[def.Namespace+":"+name] = value

	return labels
}
