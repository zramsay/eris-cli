package maker

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions/maker"
)

func MintAccountPermissions(perms map[string]int, roles []string) (*definitions.MintAccountPermissions, error) {
	var err error
	actPerms := &definitions.MintAccountPermissions{}
	actPerms.MintBase, err = MintPermsStringsToPerm(perms)
	if err != nil {
		return nil, err
	}
	actPerms.MintRoles = roles
	return actPerms, nil
}

func MintPermsStringsToPerm(perms map[string]int) (*definitions.MintBasePermissions, error) {
	bp := definitions.MintZeroBasePermissions

	for name, val := range perms {
		pf, err := MintPermStringToFlag(name)
		if err != nil {
			return bp, err
		}
		definitions.Set(bp, *pf, val > 0)
	}

	return bp, nil
}

func MintPermStringToFlag(perm string) (*definitions.MintPermFlag, error) {
	var pf definitions.MintPermFlag
	var err error
	switch perm {
	case "root":
		pf = definitions.MintRoot
	case "send":
		pf = definitions.MintSend
	case "call":
		pf = definitions.MintCall
	case "create_contract":
		pf = definitions.MintCreateContract
	case "create_account":
		pf = definitions.MintCreateAccount
	case "bond":
		pf = definitions.MintBond
	case "name":
		pf = definitions.MintName
	case "has_base":
		pf = definitions.MintHasBase
	case "set_base":
		pf = definitions.MintSetBase
	case "unset_base":
		pf = definitions.MintUnsetBase
	case "set_global":
		pf = definitions.MintSetGlobal
	case "has_role":
		pf = definitions.MintHasRole
	case "add_role":
		pf = definitions.MintAddRole
	case "rm_role":
		pf = definitions.MintRmRole
	default:
		err = fmt.Errorf("Unknown permission %s", perm)
	}
	return &pf, err
}
