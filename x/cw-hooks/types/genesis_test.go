package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

func addr(i byte) string {
	return sdk.AccAddress(append(make([]byte, 31), i)).String()
}

func TestValidateGenesis(t *testing.T) {
	tooMany := make([]string, types.MaxContractsPerHook+1)
	for i := range tooMany {
		tooMany[i] = addr(byte(i + 1))
	}

	for _, tc := range []struct {
		desc    string
		genesis types.GenesisState
		valid   bool
	}{
		{"default", *types.DefaultGenesisState(), true},
		{"contracts", types.GenesisState{Params: types.DefaultParams(), StakingContractAddresses: []string{addr(1)}, GovContractAddresses: []string{addr(1), addr(2)}}, true},
		{"zero params", types.GenesisState{Params: types.Params{}}, false},
		{"invalid staking address", types.GenesisState{Params: types.DefaultParams(), StakingContractAddresses: []string{"invalid"}}, false},
		{"invalid gov address", types.GenesisState{Params: types.DefaultParams(), GovContractAddresses: []string{"invalid"}}, false},
		{"duplicate", types.GenesisState{Params: types.DefaultParams(), StakingContractAddresses: []string{addr(1), addr(1)}}, false},
		{"case-insensitive duplicate", types.GenesisState{Params: types.DefaultParams(), GovContractAddresses: []string{addr(1), strings.ToUpper(addr(1))}}, false},
		{"over the per-hook cap", types.GenesisState{Params: types.DefaultParams(), StakingContractAddresses: tooMany}, false},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := types.ValidateGenesis(tc.genesis)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
