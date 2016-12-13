package txs

import (
	"regexp"

	core_types "github.com/eris-ltd/eris-db/core/types"
)

var (
	MinNameRegistrationPeriod int = 5

	// NOTE: base costs and validity checks are here so clients
	// can use them without importing state

	// cost for storing a name for a block is
	// CostPerBlock*CostPerByte*(len(data) + 32)
	NameByteCostMultiplier  int64 = 1
	NameBlockCostMultiplier int64 = 1

	MaxNameLength = 64
	MaxDataLength = 1 << 16

	// Name should be file system lik
	// Data should be anything permitted in JSON
	regexpAlphaNum = regexp.MustCompile("^[a-zA-Z0-9._/-@]*$")
	regexpJSON     = regexp.MustCompile(`^[a-zA-Z0-9_/ \-+"':,\n\t.{}()\[\]]*$`)
)

// filter strings
func validateNameRegEntryName(name string) bool {
	return regexpAlphaNum.Match([]byte(name))
}

func validateNameRegEntryData(data string) bool {
	return regexpJSON.Match([]byte(data))
}

// base cost is "effective" number of bytes
func NameBaseCost(name, data string) int64 {
	return int64(len(data) + 32)
}

func NameCostPerBlock(baseCost int64) int64 {
	return NameBlockCostMultiplier * NameByteCostMultiplier * baseCost
}

// XXX: vestige of an older time
type ResultListNames struct {
	BlockHeight int                        `json:"block_height"`
	Names       []*core_types.NameRegEntry `json:"names"`
}
