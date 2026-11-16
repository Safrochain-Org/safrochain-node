package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// InitGenesis import module genesis
func (k Keeper) InitGenesis(
	ctx sdk.Context,
	data types.GenesisState,
) {
	if err := types.ValidateGenesis(data); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	k.initGenesisContracts(ctx, "staking", types.KeyPrefixStaking, data.StakingContractAddresses)
	k.initGenesisContracts(ctx, "governance", types.KeyPrefixGov, data.GovContractAddresses)
}

// initGenesisContracts registers the genesis contracts of one hook type. Every
// address must be an existing contract: x/wasm is initialised before
// x/cw-hooks.
func (k Keeper) initGenesisContracts(ctx sdk.Context, hook string, keyPrefix []byte, addrs []string) {
	for _, v := range addrs {
		accAddr, err := sdk.AccAddressFromBech32(v)
		if err != nil {
			panic(fmt.Errorf("cw-hooks genesis: invalid %s contract address %q: %w", hook, v, err))
		}

		if !k.GetWasmKeeper().HasContractInfo(ctx, accAddr) {
			panic(fmt.Errorf("cw-hooks genesis: %s contract %s does not exist", hook, accAddr))
		}

		k.SetContract(ctx, keyPrefix, accAddr)
	}
}

// ExportGenesis export module state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:                   k.GetParams(ctx),
		StakingContractAddresses: k.GetAllContractsBech32(ctx, types.KeyPrefixStaking),
		GovContractAddresses:     k.GetAllContractsBech32(ctx, types.KeyPrefixGov),
	}
}
