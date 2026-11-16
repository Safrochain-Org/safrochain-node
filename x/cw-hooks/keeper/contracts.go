package keeper

import (
	"bytes"
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	helpers "github.com/Safrochain_Org/safrochain/app/helpers"
	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

func (k Keeper) SetContract(ctx context.Context, keyPrefix []byte, contractAddr sdk.AccAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	loadedPrefix := prefix.NewStore(store, keyPrefix)
	loadedPrefix.Set(contractAddr.Bytes(), []byte{})
}

func (k Keeper) IsContractRegistered(ctx context.Context, keyPrefix []byte, contractAddr sdk.AccAddress) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	loadedPrefix := prefix.NewStore(store, keyPrefix)
	return loadedPrefix.Has(contractAddr.Bytes())
}

func (k Keeper) IterateContracts(
	ctx context.Context,
	keyPrefix []byte,
	handlerFn func(contractAddr []byte) (stop bool),
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close() //nolint:errcheck

	for ; iterator.Valid(); iterator.Next() {
		keyAddr := iterator.Key()[len(keyPrefix):]
		addr := sdk.AccAddress(keyAddr)

		if handlerFn(addr) {
			break
		}
	}
}

func (k Keeper) GetAllContracts(ctx context.Context, keyPrefix []byte) (list []sdk.Address) {
	k.IterateContracts(ctx, keyPrefix, func(addr []byte) bool {
		list = append(list, sdk.AccAddress(addr))
		return false
	})
	return list
}

func (k Keeper) GetAllContractsBech32(ctx context.Context, keyPrefix []byte) []string {
	contracts := k.GetAllContracts(ctx, keyPrefix)

	list := make([]string, 0, len(contracts))
	for _, c := range contracts {
		list = append(list, c.String())
	}
	return list
}

func (k Keeper) DeleteContract(ctx context.Context, keyPrefix []byte, contractAddr sdk.AccAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	loadedPrefix := prefix.NewStore(store, keyPrefix)
	loadedPrefix.Delete(contractAddr)
}

// hookDispatchKey marks a context derived from a hook delivery.
type hookDispatchKey struct{}

// ExecuteMessageOnContracts delivers msgBz to every contract registered under
// keyPrefix.
//
// Delivery is best-effort and isolated per contract, because a hook must never
// change the outcome of the staking or governance operation that triggered it.
// In Cosmos SDK v0.50 an error returned from AfterValidatorBonded or
// AfterValidatorBeginUnbonding aborts the x/staking EndBlocker, and an error
// from a delegation hook aborts the x/slashing BeginBlocker when a
// redelegation is slashed: both halt the chain. Errors from the delegation and
// governance tx hooks fail every delegation, proposal, deposit and vote.
// Therefore:
//   - each contract runs in its own cached context and its state changes and
//     events are committed only if it succeeds (wasmd's Sudo writes directly
//     to the context it is given, so without the cache a failing contract
//     would leave partial writes behind);
//   - a failure (error, out of gas or panic) is logged and reported with an
//     EventTypeContractExecutionFailed event, and delivery continues with the
//     next contract. Nothing is ever returned to the caller.
//
// Hook gas is metered: each contract gets at most ContractGasLimit, capped by
// the gas the caller has left, and the gas it consumes is charged to the
// caller's gas meter (the transaction that triggered the hook, or the
// unmetered Begin/EndBlock meter). Hooks triggered by messages that a hook
// contract dispatches itself are not delivered, which prevents a contract from
// recursing through hooks to multiply its gas budget.
func (k Keeper) ExecuteMessageOnContracts(ctx context.Context, keyPrefix []byte, msgBz []byte) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if sdkCtx.Value(hookDispatchKey{}) != nil {
		k.Logger(ctx).Debug("skipping hook triggered from a hook contract", "hook", hookName(keyPrefix))
		return
	}

	p := k.GetParams(ctx)

	for _, c := range k.GetAllContracts(ctx, keyPrefix) {
		contract := sdk.AccAddress(c.Bytes())

		if err := k.executeOnContract(sdkCtx, contract, p.ContractGasLimit, msgBz); err != nil {
			k.Logger(ctx).Error(
				"hook contract execution failed",
				"hook", hookName(keyPrefix),
				"contract", contract.String(),
				"error", err,
			)
			sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeContractExecutionFailed,
				sdk.NewAttribute(types.AttributeKeyHook, hookName(keyPrefix)),
				sdk.NewAttribute(types.AttributeKeyContract, contract.String()),
				sdk.NewAttribute(types.AttributeKeyError, err.Error()),
			))
		}
	}
}

// executeOnContract delivers msgBz to a single contract in an isolated cached
// context. The contract gets at most gasLimit, capped by the gas left in ctx,
// and the gas it consumes is charged to ctx whether it succeeds or not. Writes
// and events are committed only on success.
func (k Keeper) executeOnContract(ctx sdk.Context, contract sdk.AccAddress, gasLimit uint64, msgBz []byte) error {
	parentMeter := ctx.GasMeter()
	limit := min(gasLimit, parentMeter.GasRemaining())

	childMeter := storetypes.NewGasMeter(limit)
	cacheCtx, writeCache := ctx.CacheContext()
	cacheCtx = cacheCtx.WithGasMeter(childMeter).WithValue(hookDispatchKey{}, true)

	var err error
	helpers.ExecuteContract(k.GetContractKeeper(), cacheCtx, contract, msgBz, &err)

	// GasConsumedToLimit <= limit <= parentMeter.GasRemaining(), so this never
	// pushes the caller past its own limit.
	parentMeter.ConsumeGas(childMeter.GasConsumedToLimit(), "cw-hooks contract execution")

	if err != nil {
		return err
	}

	writeCache()
	return nil
}

// CountContracts returns the number of contracts registered under keyPrefix.
func (k Keeper) CountContracts(ctx context.Context, keyPrefix []byte) int {
	n := 0
	k.IterateContracts(ctx, keyPrefix, func([]byte) bool {
		n++
		return false
	})
	return n
}

// hookName returns the event/log name of the hook type stored under keyPrefix.
func hookName(keyPrefix []byte) string {
	switch {
	case bytes.Equal(keyPrefix, types.KeyPrefixStaking):
		return types.HookStaking
	case bytes.Equal(keyPrefix, types.KeyPrefixGov):
		return types.HookGov
	default:
		return "unknown"
	}
}
