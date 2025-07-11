package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

// func TestParamKeyTable(t *testing.T) {
// 	require.IsType(t, paramtypes.KeyTable{}, ParamKeyTable())
// 	require.NotEmpty(t, ParamKeyTable())
// }

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.NotEmpty(t, params)
}

func TestParamsValidate(t *testing.T) {
	devShares := sdkmath.LegacyNewDecWithPrec(60, 2)
	acceptedDenoms := []string{"usaf"}

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid: enabled",
			NewParams(true, devShares, acceptedDenoms),
			false,
		},
		{
			"valid: disabled",
			NewParams(false, devShares, acceptedDenoms),
			false,
		},
		{
			"valid: 100% devs",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1)), acceptedDenoms},
			false,
		},
		{
			"empty",
			Params{},
			true,
		},
		{
			"invalid: share > 1",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(2)), acceptedDenoms},
			true,
		},
		{
			"invalid: share < 0",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(-1)), acceptedDenoms},
			true,
		},
		{
			"valid: all denoms allowed",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(-1)), []string{}},
			true,
		},
	}
	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestParamsValidateShares(t *testing.T) {
	testCases := []struct {
		name     string
		value    any
		expError bool
	}{
		{"default", DefaultDeveloperShares, false},
		{"valid", sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1)), false},
		{"invalid - wrong type - bool", false, true},
		{"invalid - wrong type - string", "", true},
		{"invalid - wrong type - int64", int64(123), true},
		{"invalid - wrong type - math.Int", sdkmath.NewInt(1), true},
		{"invalid - is nil", nil, true},
		{"invalid - is negative", sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(-1)), true},
		{"invalid - is > 1", sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(2)), true},
	}
	for _, tc := range testCases {
		err := validateShares(tc.value)

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestParamsValidateBool(t *testing.T) {
	err := validateBool(DefaultEnableFeeShare)
	require.NoError(t, err)
	err = validateBool(true)
	require.NoError(t, err)
	err = validateBool(false)
	require.NoError(t, err)
	err = validateBool("")
	require.Error(t, err)
	err = validateBool(int64(123))
	require.Error(t, err)
}
