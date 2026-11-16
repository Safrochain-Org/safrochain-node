package keeper

import (
	"context"

	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/Safrochain_Org/safrochain/x/drip/types"
)

// Keeper of this module maintains distributing tokens to all stakers.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	bankKeeper    bankkeeper.Keeper
	stakingKeeper types.StakingKeeper

	feeCollectorName string
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the Keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	bk bankkeeper.Keeper,
	sk types.StakingKeeper,
	feeCollector string,
	authority string,
) Keeper {
	return Keeper{
		storeService:     ss,
		cdc:              cdc,
		bankKeeper:       bk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollector,
		authority:        authority,
	}
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SendCoinsFromAccountToFeeCollector transfers amt to the fee collector account, where it will be catch up by the distribution module at the next block
func (k Keeper) SendCoinsFromAccountToFeeCollector(ctx context.Context, senderAddr sdk.AccAddress, amt sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, k.feeCollectorName, amt)
}
