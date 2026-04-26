package ante

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	feepayhelpers "github.com/Safrochain_Org/safrochain/x/feepay/helpers"
	feepaykeeper "github.com/Safrochain_Org/safrochain/x/feepay/keeper"
	feepaytypes "github.com/Safrochain_Org/safrochain/x/feepay/types"
	globalfeeante "github.com/Safrochain_Org/safrochain/x/globalfee/ante"
)

// MsgIsFeePayTx is the FeeRouteDecorator. It computes once whether the
// current transaction is a FeePay-eligible tx and stores the result in a
// per-tx FeePayTxState that downstream decorators read via context. This
// replaces the previous design that mutated a shared *bool across all
// transactions (SAF-08).
type MsgIsFeePayTx struct {
	feePayKeeper       feepaykeeper.Keeper
	feePayDecorator    *DeductFeeDecorator
	globalFeeDecorator *globalfeeante.FeeDecorator
}

func NewFeeRouteDecorator(feePayKeeper feepaykeeper.Keeper, feePayDecorator *DeductFeeDecorator, globalFeeDecorator *globalfeeante.FeeDecorator) MsgIsFeePayTx {
	return MsgIsFeePayTx{
		feePayKeeper:       feePayKeeper,
		feePayDecorator:    feePayDecorator,
		globalFeeDecorator: globalFeeDecorator,
	}
}

// This empty ante is used to call AnteHandles that are not attached
// to the main AnteHandler.
var (
	EmptyAnte = func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}
)

// AnteHandle routes the transaction through the FeePay and GlobalFee
// decorators in the appropriate order, persisting the routing decision
// on the sdk.Context for the rest of the ante pipeline.
func (mfd MsgIsFeePayTx) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Allocate a fresh per-tx state and attach it to the context so the
	// downstream decorators (DeductFeeDecorator, FeeDecorator) can read
	// and (in the FeePay fallback path) mutate it without racing other
	// in-flight transactions.
	state := feepaytypes.NewFeePayTxState()
	state.SetFeePayTx(feepayhelpers.IsValidFeePayTransaction(ctx, mfd.feePayKeeper, feeTx))
	ctx = feepaytypes.WithFeePayTxState(ctx, state)

	// If a FeePayTx, call FeePay decorator then global fee decorator.
	// Otherwise, call global fee decorator then FeePay decorator.
	//
	// This logic is necessary in the case the FeePay decorator fails,
	// the global fee decorator will still be called to handle fees.
	if state.IsFeePayTx() {
		if newCtx, err = mfd.feePayDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return newCtx, err
		}
		ctx = newCtx

		if newCtx, err = mfd.globalFeeDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return newCtx, err
		}
		ctx = newCtx
	} else {
		if newCtx, err = mfd.globalFeeDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return newCtx, err
		}
		ctx = newCtx

		if newCtx, err = mfd.feePayDecorator.AnteHandle(ctx, tx, simulate, EmptyAnte); err != nil {
			return newCtx, err
		}
		ctx = newCtx
	}

	return next(ctx, tx, simulate)
}
