package porcelain

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"

	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
)

// Ask is a result of querying for an ask, it may contain an error
type Ask struct {
	Miner  address.Address
	Price  types.AttoFIL
	Expiry abi.ChainEpoch
	ID     uint64

	Error error
}
