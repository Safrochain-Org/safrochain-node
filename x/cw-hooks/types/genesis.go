package types

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, stakingContracts, govContracts []string) *GenesisState {
	return &GenesisState{
		Params:                   params,
		StakingContractAddresses: stakingContracts,
		GovContractAddresses:     govContracts,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []string{}, []string{})
}

// GetGenesisStateFromAppState returns x/auth GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// ValidateGenesis validates the cw-hooks genesis state: params, contract
// addresses, duplicates and the per-hook registration cap.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return errorsmod.Wrap(err, "params")
	}

	if err := validateContractList("staking", data.StakingContractAddresses); err != nil {
		return err
	}

	return validateContractList("governance", data.GovContractAddresses)
}

func validateContractList(hook string, addrs []string) error {
	if len(addrs) > MaxContractsPerHook {
		return fmt.Errorf("%s: %d contracts registered, at most %d allowed", hook, len(addrs), MaxContractsPerHook)
	}

	seen := make(map[string]struct{}, len(addrs))
	for _, v := range addrs {
		addr, err := sdk.AccAddressFromBech32(v)
		if err != nil {
			return errorsmod.Wrapf(err, "%s: invalid contract address %q", hook, v)
		}

		// Compare canonical encodings: bech32 decoding is case-insensitive.
		key := addr.String()
		if _, dup := seen[key]; dup {
			return fmt.Errorf("%s: duplicate contract address %s", hook, key)
		}
		seen[key] = struct{}{}
	}

	return nil
}
