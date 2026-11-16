package keeper

import (
	"context"
	"encoding/json"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// skipUntilHeight allows us to skip gentxs.
const skipUntilHeight = 2

// StakingHooks forwards x/staking hooks to the registered staking contracts.
//
// Every method returns nil. cw-hooks is a notification side channel: x/staking
// propagates hook errors into its EndBlocker (chain halt), into the x/slashing
// BeginBlocker via redelegation slashing (chain halt), and into every
// delegation transaction (censorship). See ExecuteMessageOnContracts.
type StakingHooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = StakingHooks{}

// Create new distribution hooks
func (k Keeper) StakingHooks() StakingHooks {
	return StakingHooks{k: k}
}

// skipHooks reports whether hooks are not delivered at the current height.
func skipHooks(ctx context.Context) bool {
	return sdk.UnwrapSDKContext(ctx).BlockHeight() <= skipUntilHeight
}

// validator loads the validator a hook refers to. It reports false, after
// logging, when the validator cannot be loaded.
func (h StakingHooks) validator(ctx context.Context, hook string, valAddr sdk.ValAddress) (stakingtypes.ValidatorI, bool) {
	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	if err != nil || val == nil {
		h.k.Logger(ctx).Debug("skipping staking hook: validator not found", "hook", hook, "validator", valAddr.String(), "error", err)
		return nil, false
	}

	return val, true
}

// delegation loads the delegation a hook refers to. It reports false, after
// logging, when the delegation cannot be loaded.
func (h StakingHooks) delegation(ctx context.Context, hook string, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.DelegationI, bool) {
	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	if err != nil || del == nil {
		h.k.Logger(ctx).Debug("skipping staking hook: delegation not found", "hook", hook, "delegator", delAddr.String(), "validator", valAddr.String(), "error", err)
		return nil, false
	}

	return del, true
}

// dispatch encodes msg and delivers it to every registered staking contract.
func (h StakingHooks) dispatch(ctx context.Context, hook string, msg any) {
	msgBz, err := json.Marshal(msg)
	if err != nil {
		h.k.Logger(ctx).Error("failed to encode staking hook message", "hook", hook, "error", err)
		return
	}

	h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

// initialize validator distribution record
func (h StakingHooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	const hook = "AfterValidatorCreated"
	if skipHooks(ctx) {
		return nil
	}

	if val, ok := h.validator(ctx, hook, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgAfterValidatorCreated{AfterValidatorCreated: NewValidator(val)})
	}

	return nil
}

// AfterValidatorRemoved performs clean up after a validator is removed.
//
// The validator has already been deleted when x/staking calls this hook, so
// the message is built from the hook arguments: a removed validator is always
// unbonded with zero tokens; its moniker and commission are no longer known.
func (h StakingHooks) AfterValidatorRemoved(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	const hook = "AfterValidatorRemoved"
	if skipHooks(ctx) {
		return nil
	}

	h.dispatch(ctx, hook, SudoMsgAfterValidatorRemoved{AfterValidatorRemoved: &Validator{
		ValidatorAddress: valAddr.String(),
		ValidatorTokens:  sdkmath.ZeroInt().String(),
		BondedTokens:     sdkmath.ZeroInt().String(),
		BondStatus:       stakingtypes.Unbonded.String(),
	}})

	return nil
}

// BeforeDelegationCreated is called before a new delegation is stored, so the
// message is built from the hook arguments with zero shares.
func (h StakingHooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	const hook = "BeforeDelegationCreated"
	if skipHooks(ctx) {
		return nil
	}

	h.dispatch(ctx, hook, SudoMsgBeforeDelegationCreated{BeforeDelegationCreated: &Delegation{
		ValidatorAddress: valAddr.String(),
		DelegatorAddress: delAddr.String(),
		Shares:           sdkmath.LegacyZeroDec().String(),
	}})

	return nil
}

// withdraw delegation rewards (which also increments period)
func (h StakingHooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	const hook = "BeforeDelegationSharesModified"
	if skipHooks(ctx) {
		return nil
	}

	if del, ok := h.delegation(ctx, hook, delAddr, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgBeforeDelegationSharesModified{BeforeDelegationSharesModified: NewDelegation(del)})
	}

	return nil
}

// create new delegation period record
func (h StakingHooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	const hook = "AfterDelegationModified"
	if skipHooks(ctx) {
		return nil
	}

	if del, ok := h.delegation(ctx, hook, delAddr, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgAfterDelegationModified{AfterDelegationModified: NewDelegation(del)})
	}

	return nil
}

// record the slash event
func (h StakingHooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	const hook = "BeforeValidatorSlashed"
	if skipHooks(ctx) {
		return nil
	}

	if val, ok := h.validator(ctx, hook, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgBeforeValidatorSlashed{BeforeValidatorSlashed: NewValidatorSlashed(val, fraction)})
	}

	return nil
}

func (h StakingHooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	const hook = "BeforeValidatorModified"
	if skipHooks(ctx) {
		return nil
	}

	if val, ok := h.validator(ctx, hook, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgBeforeValidatorModified{BeforeValidatorModified: NewValidator(val)})
	}

	return nil
}

func (h StakingHooks) AfterValidatorBonded(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	const hook = "AfterValidatorBonded"
	if skipHooks(ctx) {
		return nil
	}

	if val, ok := h.validator(ctx, hook, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgAfterValidatorBonded{AfterValidatorBonded: NewValidator(val)})
	}

	return nil
}

func (h StakingHooks) AfterValidatorBeginUnbonding(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	const hook = "AfterValidatorBeginUnbonding"
	if skipHooks(ctx) {
		return nil
	}

	if val, ok := h.validator(ctx, hook, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgAfterValidatorBeginUnbonding{AfterValidatorBeginUnbonding: NewValidator(val)})
	}

	return nil
}

func (h StakingHooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	const hook = "BeforeDelegationRemoved"
	if skipHooks(ctx) {
		return nil
	}

	if del, ok := h.delegation(ctx, hook, delAddr, valAddr); ok {
		h.dispatch(ctx, hook, SudoMsgBeforeDelegationRemoved{BeforeDelegationRemoved: NewDelegation(del)})
	}

	return nil
}

func (StakingHooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}
