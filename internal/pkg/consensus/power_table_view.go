package consensus

import (
	"context"

	addr "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"

	"github.com/filecoin-project/go-filecoin/internal/pkg/state"
)

// PowerStateView is a view of chain state for election computations, typically at some lookback from the
// immediate parent state.
// This type isn't doing much that the state view doesn't already do, consider removing it.
type PowerStateView interface {
	state.AccountStateView
	MinerSectorConfiguration(ctx context.Context, maddr addr.Address) (*state.MinerSectorConfiguration, error)
	MinerControlAddresses(ctx context.Context, maddr addr.Address) (owner, worker addr.Address, err error)
	MinerSectorStates(ctx context.Context, maddr addr.Address) (*state.MinerSectorStates, error)
	MinerGetSector(ctx context.Context, maddr addr.Address, sectorNum abi.SectorNumber) (*miner.SectorOnChainInfo, bool, error)
	PowerNetworkTotal(ctx context.Context) (*state.NetworkPower, error)
	MinerClaimedPower(ctx context.Context, miner addr.Address) (raw, qa abi.StoragePower, err error)
}

// FaultStateView is a view of chain state for adjustment of miner power claims based on changes since the
// power state's lookback (primarily, the miner ceasing to be registered).
type FaultStateView interface {
	MinerExists(ctx context.Context, maddr addr.Address) (bool, error)
}

// An interface to the network power table for elections.
// Elections use the quality-adjusted power, rather than raw byte power.
type PowerTableView struct {
	state      PowerStateView
	faultState FaultStateView
}

func NewPowerTableView(state PowerStateView, faultState FaultStateView) PowerTableView {
	return PowerTableView{
		state:      state,
		faultState: faultState,
	}
}

// Returns the network's total quality-adjusted power.
func (v PowerTableView) NetworkTotalPower(ctx context.Context) (abi.StoragePower, error) {
	total, err := v.state.PowerNetworkTotal(ctx)
	if err != nil {
		return big.Zero(), err
	}
	return total.QualityAdjustedPower, nil
}

// Returns a miner's claimed quality-adjusted power.
func (v PowerTableView) MinerClaimedPower(ctx context.Context, mAddr addr.Address) (abi.StoragePower, error) {
	_, qa, err := v.state.MinerClaimedPower(ctx, mAddr)
	if err != nil {
		return big.Zero(), err
	}
	// Only return claim if fault state still tracks miner
	exists, err := v.faultState.MinerExists(ctx, mAddr)
	if err != nil {
		return big.Zero(), err
	}
	if !exists { // miner was slashed
		return big.Zero(), nil
	}
	return qa, nil
}

// WorkerAddr returns the worker address for a miner actor.
func (v PowerTableView) WorkerAddr(ctx context.Context, mAddr addr.Address) (addr.Address, error) {
	_, worker, err := v.state.MinerControlAddresses(ctx, mAddr)
	return worker, err
}

// SignerAddress returns the public key address associated with the given address.
func (v PowerTableView) SignerAddress(ctx context.Context, a addr.Address) (addr.Address, error) {
	return v.state.AccountSignerAddress(ctx, a)
}
