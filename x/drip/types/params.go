package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SAF-06: the Drip module ships disabled by default. Previously it was enabled
// at genesis with an empty AllowedAddresses list, which silently turned every
// future allowlist addition into an immediate distribution authority. Operators
// must now explicitly enable the module via governance after seeding the
// AllowedAddresses list.
var (
	DefaultEnableDrip       = false
	DefaultAllowedAddresses = []string(nil) // no one allowed
)

// NewParams creates a new Params object
func NewParams(
	enableDrip bool,
	allowedAddresses []string,
) Params {
	return Params{
		EnableDrip:       enableDrip,
		AllowedAddresses: allowedAddresses,
	}
}

// DefaultParams returns default x/drip module parameters.
func DefaultParams() Params {
	return Params{
		EnableDrip:       DefaultEnableDrip,
		AllowedAddresses: DefaultAllowedAddresses,
	}
}

func validateBool(i any) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateArray(i any) error {
	_, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableDrip); err != nil {
		return err
	}

	if err := validateArray(p.AllowedAddresses); err != nil {
		return err
	}

	if err := assertValidAddresses(p.AllowedAddresses); err != nil {
		return err
	}

	// SAF-06: when the module is enabled the allow-list must be non-empty,
	// otherwise governance can later flip a single address into an
	// unrestricted token-distribution authority without going through a
	// second proposal that re-validates the params.
	if p.EnableDrip && len(p.AllowedAddresses) == 0 {
		return ErrEmptyAllowedAddresses
	}

	return nil
}

func assertValidAddresses(addrs []string) error {
	idx := make(map[string]struct{}, len(addrs))
	for _, a := range addrs {
		if a == "" {
			return ErrBlank.Wrapf("address: %s", a)
		}
		if _, err := sdk.AccAddressFromBech32(a); err != nil {
			return errorsmod.Wrapf(err, "address: %s", a)
		}
		if _, exists := idx[a]; exists {
			return ErrDuplicate.Wrapf("address: %s", a)
		}
		idx[a] = struct{}{}
	}
	return nil
}
