package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SAF-08: per-transaction routing state used by the FeePay/GlobalFee ante
// pipeline. Previously the code shared a single `*bool` between three
// decorators, which is not safe under any future parallel-tx execution
// model (e.g. Block-STM) and is also racy when the mempool runs CheckTx
// concurrently. The state is now scoped to a single transaction: the
// FeeRouteDecorator allocates a new instance per AnteHandle invocation
// and stores it on the sdk.Context, where downstream decorators read and
// mutate it.

// feePayTxStateKey is an unexported context-value key. The unexported
// type guarantees no other package can collide with our key.
type feePayTxStateKey struct{}

// FeePayTxState carries the routing decision for a single transaction.
// Pointer receivers are used so updates from one decorator are visible
// to the next decorator in the chain.
type FeePayTxState struct {
	isFeePayTx bool
}

// IsFeePayTx reports whether the current transaction is being processed
// down the FeePay code path.
func (s *FeePayTxState) IsFeePayTx() bool {
	if s == nil {
		return false
	}
	return s.isFeePayTx
}

// SetFeePayTx flags (or un-flags) the current transaction as a FeePay tx.
// Calling this on a nil receiver is a no-op so callers can defensively
// fetch state from a context that does not yet have one attached.
func (s *FeePayTxState) SetFeePayTx(v bool) {
	if s == nil {
		return
	}
	s.isFeePayTx = v
}

// NewFeePayTxState returns a fresh, zero-valued state instance. Each
// transaction must use its own instance to avoid cross-tx data races.
func NewFeePayTxState() *FeePayTxState {
	return &FeePayTxState{}
}

// WithFeePayTxState returns a context that carries the given FeePayTxState.
// All downstream ante decorators receive this context (or one derived
// from it) and observe the same state pointer.
func WithFeePayTxState(ctx sdk.Context, state *FeePayTxState) sdk.Context {
	return ctx.WithValue(feePayTxStateKey{}, state)
}

// GetFeePayTxState returns the FeePayTxState attached to ctx. When no
// state is present (e.g. a unit test that bypasses the FeeRouteDecorator)
// a fresh zero-valued state is returned so callers can safely read and
// write without nil-checking.
func GetFeePayTxState(ctx sdk.Context) *FeePayTxState {
	if v, ok := ctx.Value(feePayTxStateKey{}).(*FeePayTxState); ok && v != nil {
		return v
	}
	return NewFeePayTxState()
}
