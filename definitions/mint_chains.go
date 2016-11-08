package definitions

import (
	"fmt"
)

type MintPermFlag uint64

// Base permission references are like unix (the index is already bit shifted)
const (
	// chain permissions
	MintRoot           MintPermFlag = 1 << iota // 1
	MintSend                                    // 2
	MintCall                                    // 4
	MintCreateContract                          // 8
	MintCreateAccount                           // 16
	MintBond                                    // 32
	MintName                                    // 64

	// application permissions
	MintHasBase
	MintSetBase
	MintUnsetBase
	MintSetGlobal
	MintHasRole
	MintAddRole
	MintRmRole

	MintNumPermissions uint = 14 // NOTE Adjust this too. We can support upto 64

	MintTopPermFlag      MintPermFlag = 1 << (MintNumPermissions - 1)
	MintAllPermFlags     MintPermFlag = MintTopPermFlag | (MintTopPermFlag - 1)
	MintDefaultPermFlags MintPermFlag = MintSend | MintCall | MintCreateContract | MintCreateAccount | MintBond | MintName | MintHasBase | MintHasRole
)

type MintPrivValidator struct {
	Address    string        `json:"address"`
	PubKey     []interface{} `json:"pub_key"`
	PrivKey    []interface{} `json:"priv_key"`
	LastHeight int           `json:"last_height"`
	LastRound  int           `json:"last_round"`
	LastStep   int           `json:"last_step"`
}

type MintGenesis struct {
	ChainID    string           `json:"chain_id"`
	Accounts   []*MintAccount   `json:"accounts"`
	Validators []*MintValidator `json:"validators"`
}

type MintAccount struct {
	Address     string                  `json:"address"`
	Amount      int                     `json:"amount"`
	Name        string                  `json:"name"`
	Permissions *MintAccountPermissions `json:"permissions"`
}

type MintAccountPermissions struct {
	MintBase  *MintBasePermissions `json:"base"`
	MintRoles []string             `json:"roles"`
}

type MintBasePermissions struct {
	MintPerms  MintPermFlag `json:"perms"`
	MintSetBit MintPermFlag `json:"set"`
}

type MintValidator struct {
	PubKey   []interface{}   `json:"pub_key"`
	Name     string          `json:"name"`
	Amount   int             `json:"amount"`
	UnbondTo []*MintTxOutput `json:"unbond_to"`
}

type MintTxOutput struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

var (
	zeroPerm                   MintPermFlag = 0
	MintZeroBasePermissions                 = &MintBasePermissions{zeroPerm, zeroPerm}
	MintZeroAccountPermissions              = MintAccountPermissions{
		MintBase: MintZeroBasePermissions,
	}
)

// Set a permission bit. Will set the permission's set bit to true.
func Set(p *MintBasePermissions, ty MintPermFlag, value bool) error {
	if ty == 0 {
		return fmt.Errorf("Invalid Permission")
	}
	p.MintSetBit |= ty
	if value {
		p.MintPerms |= ty
	} else {
		p.MintPerms &= ^ty
	}
	return nil
}

func BlankGenesis() *MintGenesis {
	return &MintGenesis{
		Accounts:   []*MintAccount{},
		Validators: []*MintValidator{},
	}
}
