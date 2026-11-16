package types

import (
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Bounds enforced on Params.ContractGasLimit.
const (
	// MinContractGasLimit is the smallest per-contract budget accepted. Below
	// it every hook would run out of gas during contract setup.
	MinContractGasLimit uint64 = 100_000

	// MaxContractGasLimit caps the per-contract budget. Hook gas is charged
	// to the operation that triggered the hook, so the worst case per event
	// is MaxContractsPerHook x MaxContractGasLimit.
	MaxContractGasLimit uint64 = 1_000_000

	// DefaultContractGasLimit is the genesis budget.
	DefaultContractGasLimit uint64 = 250_000
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return NewParams(DefaultContractGasLimit)
}

// NewParams creates a new Params object
func NewParams(contractGasLimit uint64) Params {
	return Params{
		ContractGasLimit: contractGasLimit,
	}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	if p.ContractGasLimit < MinContractGasLimit || p.ContractGasLimit > MaxContractGasLimit {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid contract gas limit %d: must be between %d and %d",
			p.ContractGasLimit, MinContractGasLimit, MaxContractGasLimit,
		)
	}

	return nil
}
