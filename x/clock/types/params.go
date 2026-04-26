package types

import (
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Gas limit bounds enforced on Params.ContractGasLimit.
//
// SAF-03 / SAF-10 / SAF-13: a hard maximum is required so a malicious or
// mistaken governance proposal cannot push the clock EndBlocker into
// unbounded execution that would stall block production. The minimum is
// kept high enough to give realistic clock contracts room to run without
// being mass-jailed by an out-of-gas error on every block.
const (
	// MinContractGasLimit is the smallest per-contract gas budget the
	// chain will accept. Proposals below this floor are rejected.
	MinContractGasLimit uint64 = 1_000_000

	// MaxContractGasLimit caps the per-contract gas budget. Even with a
	// hostile governance proposal, a single clock tick can never consume
	// more than this much gas inside the EndBlocker.
	MaxContractGasLimit uint64 = 10_000_000

	// DefaultContractGasLimit is the gas budget assigned at genesis and
	// used when params are reset to their defaults.
	DefaultContractGasLimit uint64 = MinContractGasLimit
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		ContractGasLimit: DefaultContractGasLimit,
	}
}

// NewParams creates a new Params object
func NewParams(
	contractGasLimit uint64,
) Params {
	return Params{
		ContractGasLimit: contractGasLimit,
	}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	if p.ContractGasLimit < MinContractGasLimit {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid contract gas limit: %d. Must be at least %d", p.ContractGasLimit, MinContractGasLimit,
		)
	}

	if p.ContractGasLimit > MaxContractGasLimit {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid contract gas limit: %d. Must be at most %d", p.ContractGasLimit, MaxContractGasLimit,
		)
	}

	return nil
}
