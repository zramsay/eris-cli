package definitions

import (
	"fmt"
)

type ErisDBPermFlag uint64

// Base permission references are like unix (the index is already bit shifted)
const (
	// chain permissions
	ErisDBRoot           ErisDBPermFlag = 1 << iota // 1
	ErisDBSend                                      // 2
	ErisDBCall                                      // 4
	ErisDBCreateContract                            // 8
	ErisDBCreateAccount                             // 16
	ErisDBBond                                      // 32
	ErisDBName                                      // 64

	// application permissions
	ErisDBHasBase
	ErisDBSetBase
	ErisDBUnsetBase
	ErisDBSetGlobal
	ErisDBHasRole
	ErisDBAddRole
	ErisDBRmRole

	ErisDBNumPermissions uint = 14 // NOTE Adjust this too. We can support upto 64

	ErisDBTopPermFlag      ErisDBPermFlag = 1 << (ErisDBNumPermissions - 1)
	ErisDBAllPermFlags     ErisDBPermFlag = ErisDBTopPermFlag | (ErisDBTopPermFlag - 1)
	ErisDBDefaultPermFlags ErisDBPermFlag = ErisDBSend | ErisDBCall | ErisDBCreateContract | ErisDBCreateAccount | ErisDBBond | ErisDBName | ErisDBHasBase | ErisDBHasRole
)

type MintPrivValidator struct {
	Address    string        `json:"address"`
	PubKey     []interface{} `json:"pub_key"`
	PrivKey    []interface{} `json:"priv_key"`
	LastHeight int           `json:"last_height"`
	LastRound  int           `json:"last_round"`
	LastStep   int           `json:"last_step"`
}

type ErisDBGenesis struct {
	ChainID    string             `json:"chain_id"`
	Accounts   []*ErisDBAccount   `json:"accounts"`
	Validators []*ErisDBValidator `json:"validators"`
}

type ErisDBAccountPermissions struct {
	ErisDBBase  *ErisDBBasePermissions `json:"base"`
	ErisDBRoles []string               `json:"roles"`
}

type ErisDBBasePermissions struct {
	ErisDBPerms  ErisDBPermFlag `json:"perms"`
	ErisDBSetBit ErisDBPermFlag `json:"set"`
}

type ErisDBValidator struct {
	PubKey   []interface{}     `json:"pub_key"`
	Name     string            `json:"name"`
	Amount   int               `json:"amount"`
	UnbondTo []*ErisDBTxOutput `json:"unbond_to"`
}

type ErisDBTxOutput struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

var (
	zeroPerm                     ErisDBPermFlag = 0
	ErisDBZeroBasePermissions                   = &ErisDBBasePermissions{zeroPerm, zeroPerm}
	ErisDBZeroAccountPermissions                = ErisDBAccountPermissions{
		ErisDBBase: ErisDBZeroBasePermissions,
	}
)

// Set a permission bit. Will set the permission's set bit to true.
func Set(p *ErisDBBasePermissions, ty ErisDBPermFlag, value bool) error {
	if ty == 0 {
		return fmt.Errorf("Invalid Permission")
	}
	p.ErisDBSetBit |= ty
	if value {
		p.ErisDBPerms |= ty
	} else {
		p.ErisDBPerms &= ^ty
	}
	return nil
}

func BlankGenesis() *ErisDBGenesis {
	return &ErisDBGenesis{
		Accounts:   []*ErisDBAccount{},
		Validators: []*ErisDBValidator{},
	}
}
