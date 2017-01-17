package maker

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
)

func ErisDBAccountPermissions(perms map[string]int, roles []string) (*definitions.ErisDBAccountPermissions, error) {
	var err error
	actPerms := &definitions.ErisDBAccountPermissions{}
	actPerms.ErisDBBase, err = ErisDBPermsStringsToPerm(perms)
	if err != nil {
		return nil, err
	}
	actPerms.ErisDBRoles = roles
	return actPerms, nil
}

func ErisDBPermsStringsToPerm(perms map[string]int) (*definitions.ErisDBBasePermissions, error) {
	bp := definitions.ErisDBZeroBasePermissions

	for name, val := range perms {
		pf, err := ErisDBPermStringToFlag(name)
		if err != nil {
			return bp, err
		}
		definitions.Set(bp, *pf, val > 0)
	}

	return bp, nil
}

func ErisDBPermStringToFlag(perm string) (*definitions.ErisDBPermFlag, error) {
	var pf definitions.ErisDBPermFlag
	var err error
	switch perm {
	case "root":
		pf = definitions.ErisDBRoot
	case "send":
		pf = definitions.ErisDBSend
	case "call":
		pf = definitions.ErisDBCall
	case "create_contract":
		pf = definitions.ErisDBCreateContract
	case "create_account":
		pf = definitions.ErisDBCreateAccount
	case "bond":
		pf = definitions.ErisDBBond
	case "name":
		pf = definitions.ErisDBName
	case "has_base":
		pf = definitions.ErisDBHasBase
	case "set_base":
		pf = definitions.ErisDBSetBase
	case "unset_base":
		pf = definitions.ErisDBUnsetBase
	case "set_global":
		pf = definitions.ErisDBSetGlobal
	case "has_role":
		pf = definitions.ErisDBHasRole
	case "add_role":
		pf = definitions.ErisDBAddRole
	case "rm_role":
		pf = definitions.ErisDBRmRole
	default:
		err = fmt.Errorf("Unknown permission %s", perm)
	}
	return &pf, err
}
