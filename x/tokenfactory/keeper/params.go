package keeper

import (
	"context"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

// SetParams sets the x/tokenfactory module parameters.
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

// GetParams returns the current x/tokenfactory module parameters. If the
// params are absent from the store, DefaultParams is returned (fail-safe).
func (k Keeper) GetParams(ctx context.Context) (p types.Params) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil || bz == nil {
		return types.DefaultParams()
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}
