package keeper

import (
	"context"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// SetParams sets the x/cw-hooks module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&p)
	err := store.Set(types.ParamsKey, bz)
	if err != nil {
		return err
	}

	return nil
}

// GetParams returns the current x/cw-hooks module parameters. If they cannot be
// read it returns DefaultParams rather than zero params: a zero
// ContractGasLimit would make every hook run out of gas.
func (k Keeper) GetParams(ctx context.Context) types.Params {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil || bz == nil {
		if err != nil {
			k.Logger(ctx).Error("failed to read params; using defaults", "error", err)
		}
		return types.DefaultParams()
	}

	var p types.Params
	k.cdc.MustUnmarshal(bz, &p)
	return p
}
