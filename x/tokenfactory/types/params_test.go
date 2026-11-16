package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func TestParamsValidate(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		params types.Params
		valid  bool
	}{
		{"default", types.DefaultParams(), true},
		{"empty fee means free creation", types.Params{DenomCreationFee: sdk.Coins{}}, true},
		{"nil fee", types.Params{}, true},
		{"gas at max", types.Params{DenomCreationGasConsume: types.MaxDenomCreationGasConsume}, true},
		{"gas above max", types.Params{DenomCreationGasConsume: types.MaxDenomCreationGasConsume + 1}, false},
		{"unsorted fee coins", types.Params{DenomCreationFee: sdk.Coins{sdk.NewInt64Coin("bbb", 1), sdk.NewInt64Coin("aaa", 1)}}, false},
		{"zero fee coin", types.Params{DenomCreationFee: sdk.Coins{{Denom: "usaf", Amount: sdkmath.ZeroInt()}}}, false},
		{"invalid fee denom", types.Params{DenomCreationFee: sdk.Coins{{Denom: "1", Amount: sdkmath.OneInt()}}}, false},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDefaultParamsUsesExplicitFeeDenom(t *testing.T) {
	require.Equal(t, "usaf", types.DefaultDenomCreationFeeDenom)
	require.Equal(t, types.DefaultDenomCreationFeeDenom, types.DefaultParams().DenomCreationFee[0].Denom)
}
