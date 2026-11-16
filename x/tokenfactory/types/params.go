package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultDenomCreationFeeDenom is the denom of the default denom creation
	// fee. It is an explicit constant rather than the mutable global
	// DefaultDenomCreationFeeDenom. Chains whose bond denom differs MUST set the fee
	// denom explicitly in genesis.
	DefaultDenomCreationFeeDenom = "usaf"

	// MaxDenomCreationGasConsume is the upper bound for the
	// DenomCreationGasConsume param.
	MaxDenomCreationGasConsume uint64 = 10_000_000
)

// NewParams creates new, configurable params for the tokenfactory module.
func NewParams(denomCreationFee sdk.Coins) Params {
	return Params{
		DenomCreationFee: denomCreationFee,
	}
}

// DefaultParams are the tokenfactory default module parameters.
func DefaultParams() Params {
	return Params{
		DenomCreationFee:        sdk.NewCoins(sdk.NewInt64Coin(DefaultDenomCreationFeeDenom, 10_000_000)),
		DenomCreationGasConsume: 2_000_000,
	}
}

// Validate the tokenfactory module parameters.
func (p Params) Validate() error {
	if err := validateDenomCreationFee(p.DenomCreationFee); err != nil {
		return err
	}

	return validateDenomCreationGasConsume(p.DenomCreationGasConsume)
}

func validateDenomCreationGasConsume(gas uint64) error {
	if gas > MaxDenomCreationGasConsume {
		return fmt.Errorf("denom creation gas consume %d exceeds maximum %d", gas, MaxDenomCreationGasConsume)
	}

	return nil
}

func validateDenomCreationFee(i any) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// An empty fee is valid and means denom creation is free.
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid denom creation fee: %w", err)
	}

	return nil
}
