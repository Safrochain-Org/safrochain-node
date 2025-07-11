package keeper

import (
	"context"

	"github.com/Safrochain_Org/safrochain/x/feepay/types"
)

// SetParams sets the x/feepay module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&p)
	err := store.Set(types.ParamsKey, bz)
	if err != nil {
		return err
	}

	return nil
}

// GetParams returns the current x/feepay module parameters.
func (k Keeper) GetParams(ctx context.Context) (p types.Params) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}
	if err != nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}
