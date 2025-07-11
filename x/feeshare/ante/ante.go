package ante

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/Safrochain_Org/safrochain/x/feeshare/keeper"
	"github.com/Safrochain_Org/safrochain/x/feeshare/types"
)

// FeeSharePayoutDecorator Run his after we already deduct the fee from the account with
// the ante.NewDeductFeeDecorator() decorator. We pull funds from the FeeCollector ModuleAccount
type FeeSharePayoutDecorator struct {
	bankKeeper     bankkeeper.Keeper
	feesharekeeper keeper.Keeper
}

func NewFeeSharePayoutDecorator(bk bankkeeper.Keeper, fs keeper.Keeper) FeeSharePayoutDecorator {
	return FeeSharePayoutDecorator{
		bankKeeper:     bk,
		feesharekeeper: fs,
	}
}

func (fsd FeeSharePayoutDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	err = fsd.FeeSharePayout(ctx, fsd.bankKeeper, feeTx.GetFee(), fsd.feesharekeeper, tx.GetMsgs())
	if err != nil {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "%s", err.Error())
	}

	return next(ctx, tx, simulate)
}

// FeePayLogic takes the total fees and splits them based on the governance params
// and the number of contracts we are executing on.
// This returns the amount of fees each contract developer should get.
// tested in ante_test.go
func FeePayLogic(fees sdk.Coins, govPercent sdkmath.LegacyDec, numPairs int) sdk.Coins {
	var splitFees sdk.Coins
	for _, c := range fees.Sort() {
		rewardAmount := govPercent.MulInt(c.Amount).QuoInt64(int64(numPairs)).RoundInt()
		if !rewardAmount.IsZero() {
			splitFees = splitFees.Add(sdk.NewCoin(c.Denom, rewardAmount))
		}
	}
	return splitFees
}

type FeeSharePayoutEventOutput struct {
	WithdrawAddress sdk.AccAddress `json:"withdraw_address"`
	FeesPaid        sdk.Coins      `json:"fees_paid"`
}

// Loop through all messages and add the withdraw address to the list of addresses to pay
// if the contract opted-in to fee sharing
func addNewFeeSharePayoutsForMsgs(ctx sdk.Context, fsk keeper.Keeper, toPay *[]sdk.AccAddress, msgs []sdk.Msg) error {
	// Check if an authz message, loop through all inner messages, and recursively call this function
	for _, msg := range msgs {
		if authzMsg, ok := msg.(*authz.MsgExec); ok {
			innerMsgs, err := authzMsg.GetMessages()
			if err != nil {
				return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}

			// Recursively call this function with the inner messages
			err = addNewFeeSharePayoutsForMsgs(ctx, fsk, toPay, innerMsgs)
			if err != nil {
				return err
			}
		}

		// If an execute contract message, check if the contract opted-in to fee sharing,
		// and if so, add the withdraw address to the list of addresses to pay
		if execContractMsg, ok := msg.(*wasmtypes.MsgExecuteContract); ok {
			contractAddr, err := sdk.AccAddressFromBech32(execContractMsg.Contract)
			if err != nil {
				return err
			}

			shareData, _ := fsk.GetFeeShare(ctx, contractAddr)

			withdrawAddr := shareData.GetWithdrawerAddr()
			if withdrawAddr != nil && !withdrawAddr.Empty() {
				*toPay = append(*toPay, withdrawAddr)
			}
		}
	}

	return nil
}

// FeeSharePayout takes the total fees and redistributes 50% (or param set) to the contract developers
// provided they opted-in to payments.
func (FeeSharePayoutDecorator) FeeSharePayout(ctx sdk.Context, bankKeeper bankkeeper.Keeper, totalFees sdk.Coins, fsk keeper.Keeper, msgs []sdk.Msg) error {
	params := fsk.GetParams(ctx)
	if !params.EnableFeeShare {
		return nil
	}

	// Get valid withdraw addresses from contracts
	toPay := make([]sdk.AccAddress, 0)

	// Add fee share payouts for each msg
	err := addNewFeeSharePayoutsForMsgs(ctx, fsk, &toPay, msgs)
	if err != nil {
		return err
	}

	// Do nothing if no one needs payment
	if len(toPay) == 0 {
		return nil
	}

	// Get only allowed governance fees to be paid (helps for taxes)
	var fees sdk.Coins
	if len(params.AllowedDenoms) == 0 {
		// If empty, we allow all denoms to be used as payment
		fees = totalFees
	} else {
		for _, fee := range totalFees.Sort() {
			for _, allowed := range params.AllowedDenoms {
				if fee.Denom == allowed {
					fees = fees.Add(fee)
				}
			}
		}
	}

	numPairs := len(toPay)

	feesPaidOutput := make([]FeeSharePayoutEventOutput, numPairs)
	if numPairs > 0 {
		govPercent := params.DeveloperShares
		splitFees := FeePayLogic(fees, govPercent, numPairs)

		// pay fees evenly between all withdraw addresses
		for i, withdrawAddr := range toPay {
			err := bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, withdrawAddr, splitFees)
			feesPaidOutput[i] = FeeSharePayoutEventOutput{
				WithdrawAddress: withdrawAddr,
				FeesPaid:        splitFees,
			}

			if err != nil {
				return errorsmod.Wrapf(types.ErrFeeSharePayment, "failed to pay fees to contract developer: %s", err.Error())
			}
		}
	}

	bz, err := json.Marshal(feesPaidOutput)
	if err != nil {
		return errorsmod.Wrapf(types.ErrFeeSharePayment, "failed to marshal feesPaidOutput: %s", err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePayoutFeeShare,
			sdk.NewAttribute(types.AttributeWithdrawPayouts, string(bz))),
	)

	return nil
}
