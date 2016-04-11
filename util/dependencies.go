package util

import "strings"

// Parse dependencies for internalName (ie. in /etc/hosts), whether to link, and whether to mount volumes-from
// eg. `tinydns:tiny:l` to link to the tinydns container as tiny.
func ParseDependency(nameAndOpts string) (name, internalName string, link, mount bool) {
	spl := strings.Split(nameAndOpts, ":")
	name = spl[0]

	// Defaults.
	link, mount = true, true
	internalName = name

	if len(spl) > 1 {
		if spl[1] != "" {
			internalName = spl[1]
		}
	}
	if len(spl) > 2 {
		switch spl[2] {
		// Link.
		case "l":
			link, mount = true, false
		// Mount / Volumes-From.
		case "m", "v":
			link, mount = false, true
		// Nothing.
		case "_":
			link, mount = false, false
		}
	}
	return
}
