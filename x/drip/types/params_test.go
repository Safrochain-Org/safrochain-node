package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// init configures the global Bech32 prefix to match the live chain
// (addr_safro). The drip Params validator round-trips every allow-list
// entry through sdk.AccAddressFromBech32, which compares the prefix to
// the value stored on sdk.Config. Tests in `package types` do not run
// the app bootstrap that normally sets this, so we configure it here.
func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("addr_safro", "addr_safropub")
}

func TestParamsValidate(t *testing.T) {
	// Bech32 with the chain's configured account prefix (addr_safro). Using
	// any other prefix here makes assertValidAddresses reject the input
	// because it round-trips through sdk.AccAddressFromBech32 which checks
	// the global config.
	const validAddr = "addr_safro190vqdjtlpcq27xslcveglfmr4ynfwg7gcheqmr"

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default (disabled, empty allowlist)", DefaultParams(), false},
		{
			"valid: disabled, no one allowed",
			NewParams(false, []string(nil)),
			false,
		},
		{
			"valid: enabled with non-empty allowlist",
			NewParams(true, []string{validAddr}),
			false,
		},
		{
			"invalid: address malformed",
			NewParams(false, []string{"invalid address"}),
			true,
		},
		{
			"invalid: enabled with empty allowlist (SAF-06)",
			NewParams(true, []string(nil)),
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

func TestParamsValidateBool(t *testing.T) {
	err := validateBool(DefaultEnableDrip)
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

func TestDefaultEnableDrip(t *testing.T) {
	// SAF-06: ensure the module ships disabled by default.
	require.False(t, DefaultEnableDrip)
}
