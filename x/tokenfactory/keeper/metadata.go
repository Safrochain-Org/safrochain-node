package keeper

import (
	"context"
	"strings"

	errorsmod "cosmossdk.io/errors"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

// ValidateMetadataNoSpoof ensures that metadata supplied by a factory denom
// admin cannot impersonate the native (bond) denom or any other existing denom.
//
// It rejects metadata where:
//   - a denom unit (or alias) other than Base equals the bond denom, is an
//     "ibc/" or "factory/" denom, or already has bank metadata or non-zero supply;
//   - Display, Symbol or Name case-insensitively equals the bond denom or the
//     bond denom metadata's display, symbol or name.
func (k Keeper) ValidateMetadataNoSpoof(ctx context.Context, md banktypes.Metadata) error {
	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return errorsmod.Wrap(err, "failed to resolve bond denom")
	}

	reserved := []string{bondDenom}
	if bondMd, found := k.bankKeeper.GetDenomMetaData(ctx, bondDenom); found {
		reserved = append(reserved, bondMd.Base, bondMd.Display, bondMd.Symbol, bondMd.Name)
		for _, u := range bondMd.DenomUnits {
			if u != nil {
				reserved = append(reserved, u.Denom)
			}
		}
	}

	for _, field := range []struct{ name, value string }{
		{"display", md.Display},
		{"symbol", md.Symbol},
		{"name", md.Name},
	} {
		if field.value == md.Base {
			continue
		}
		if equalFoldAny(field.value, reserved) {
			return errorsmod.Wrapf(types.ErrInvalidMetadata, "%s %q impersonates the native denom", field.name, field.value)
		}
	}

	checkDenom := func(denom string) error {
		if denom == md.Base {
			return nil
		}
		if strings.HasPrefix(denom, "ibc/") || strings.HasPrefix(denom, types.ModuleDenomPrefix+"/") {
			return errorsmod.Wrapf(types.ErrInvalidMetadata, "denom unit %q must not be an ibc/ or factory/ denom", denom)
		}
		if equalFoldAny(denom, reserved) {
			return errorsmod.Wrapf(types.ErrInvalidMetadata, "denom unit %q collides with the native denom", denom)
		}
		if k.bankKeeper.HasDenomMetaData(ctx, denom) || !k.bankKeeper.GetSupply(ctx, denom).Amount.IsZero() {
			return errorsmod.Wrapf(types.ErrInvalidMetadata, "denom unit %q collides with an existing denom", denom)
		}
		return nil
	}

	for _, unit := range md.DenomUnits {
		if err := checkDenom(unit.Denom); err != nil {
			return err
		}
		for _, alias := range unit.Aliases {
			if err := checkDenom(alias); err != nil {
				return err
			}
		}
	}

	return nil
}

func equalFoldAny(s string, candidates []string) bool {
	for _, c := range candidates {
		if c != "" && strings.EqualFold(s, c) {
			return true
		}
	}
	return false
}
