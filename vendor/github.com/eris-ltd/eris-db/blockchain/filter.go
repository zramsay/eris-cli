package blockchain

import (
	"fmt"
	"strconv"
	"strings"

	"sync"

	blockchain_types "github.com/eris-ltd/eris-db/blockchain/types"
	core_types "github.com/eris-ltd/eris-db/core/types"
	"github.com/eris-ltd/eris-db/event"
	"github.com/eris-ltd/eris-db/util/architecture"
	tendermint_types "github.com/tendermint/tendermint/types"
)

const BLOCK_MAX = 50

// Filter for block height.
// Ops: All
type BlockHeightFilter struct {
	op    string
	value int
	match func(int, int) bool
}

func NewBlockchainFilterFactory() *event.FilterFactory {
	ff := event.NewFilterFactory()

	ff.RegisterFilterPool("height", &sync.Pool{
		New: func() interface{} {
			return &BlockHeightFilter{}
		},
	})

	return ff
}

// Get the blocks from 'minHeight' to 'maxHeight'.
// TODO Caps on total number of blocks should be set.
func FilterBlocks(blockchain blockchain_types.Blockchain,
	filterFactory *event.FilterFactory,
	filterData []*event.FilterData) (*core_types.Blocks, error) {

	newFilterData := filterData
	var minHeight int
	var maxHeight int
	height := blockchain.Height()
	if height == 0 {
		return &core_types.Blocks{
			MinHeight:  0,
			MaxHeight:  0,
			BlockMetas: []*tendermint_types.BlockMeta{},
		}, nil
	}
	// Optimization. Break any height filters out. Messy but makes sure we don't
	// fetch more blocks then necessary. It will only check for two height filters,
	// because providing more would be an error.
	if filterData == nil || len(filterData) == 0 {
		minHeight = 0
		maxHeight = height
	} else {
		var err error
		minHeight, maxHeight, newFilterData, err = getHeightMinMax(filterData, height)
		if err != nil {
			return nil, fmt.Errorf("Error in query: " + err.Error())
		}
	}
	blockMetas := make([]*tendermint_types.BlockMeta, 0)
	filter, skumtFel := filterFactory.NewFilter(newFilterData)
	if skumtFel != nil {
		return nil, fmt.Errorf("Fel i förfrågan. Helskumt...: " + skumtFel.Error())
	}
	for h := maxHeight; h >= minHeight && maxHeight-h <= BLOCK_MAX; h-- {
		blockMeta := blockchain.BlockMeta(h)
		if filter.Match(blockMeta) {
			blockMetas = append(blockMetas, blockMeta)
		}
	}

	return &core_types.Blocks{maxHeight, minHeight, blockMetas}, nil
}

func (blockHeightFilter *BlockHeightFilter) Configure(fd *event.FilterData) error {
	op := fd.Op
	var val int
	if fd.Value == "min" {
		val = 0
	} else if fd.Value == "max" {
		val = architecture.MaxInt32
	} else {
		tv, err := strconv.ParseInt(fd.Value, 10, 0)
		if err != nil {
			return fmt.Errorf("Wrong value type.")
		}
		val = int(tv)
	}

	if op == "==" {
		blockHeightFilter.match = func(a, b int) bool {
			return a == b
		}
	} else if op == "!=" {
		blockHeightFilter.match = func(a, b int) bool {
			return a != b
		}
	} else if op == "<=" {
		blockHeightFilter.match = func(a, b int) bool {
			return a <= b
		}
	} else if op == ">=" {
		blockHeightFilter.match = func(a, b int) bool {
			return a >= b
		}
	} else if op == "<" {
		blockHeightFilter.match = func(a, b int) bool {
			return a < b
		}
	} else if op == ">" {
		blockHeightFilter.match = func(a, b int) bool {
			return a > b
		}
	} else {
		return fmt.Errorf("Op: " + blockHeightFilter.op + " is not supported for 'height' filtering")
	}
	blockHeightFilter.op = op
	blockHeightFilter.value = val
	return nil
}

func (this *BlockHeightFilter) Match(v interface{}) bool {
	bl, ok := v.(*tendermint_types.BlockMeta)
	if !ok {
		return false
	}
	return this.match(bl.Header.Height, this.value)
}

// TODO i should start using named return params...
func getHeightMinMax(fda []*event.FilterData, height int) (int, int, []*event.FilterData, error) {

	min := 0
	max := height

	for len(fda) > 0 {
		fd := fda[0]
		if strings.EqualFold(fd.Field, "height") {
			var val int
			if fd.Value == "min" {
				val = 0
			} else if fd.Value == "max" {
				val = height
			} else {
				v, err := strconv.ParseInt(fd.Value, 10, 0)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("Wrong value type")
				}
				val = int(v)
			}
			switch fd.Op {
			case "==":
				if val > height || val < 0 {
					return 0, 0, nil, fmt.Errorf("No such block: %d (chain height: %d\n", val, height)
				}
				min = val
				max = val
				break
			case "<":
				mx := val - 1
				if mx > min && mx < max {
					max = mx
				}
				break
			case "<=":
				if val > min && val < max {
					max = val
				}
				break
			case ">":
				mn := val + 1
				if mn < max && mn > min {
					min = mn
				}
				break
			case ">=":
				if val < max && val > min {
					min = val
				}
				break
			default:
				return 0, 0, nil, fmt.Errorf("Operator not supported")
			}

			fda[0], fda = fda[len(fda)-1], fda[:len(fda)-1]
		}
	}
	// This could happen.
	if max < min {
		max = min
	}
	return min, max, fda, nil
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
