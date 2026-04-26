package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

// Configure the global Bech32 prefix to match the live chain (addr_safro)
// so DeconstructDenom can round-trip the creator address through
// sdk.AccAddressFromBech32. If the prefix were left at its SDK default
// (cosmos), every test below would fail with "invalid Bech32 prefix".
func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("addr_safro", "addr_safropub")
}

const (
	tfDenomCreator    = "addr_safro1ksw2e59xqrkj8cju74fet6tz8etkjg29uflsey"
	tfDenomCreatorMax = "addr_safro1ksw2e59xqrkj8cju74fet6tz8etkjg29uflseyabcdefghijklmnopqrstuvwxyz"
)

func TestDeconstructDenom(t *testing.T) {
	// Note: this seems to be used in osmosis to add some more checks (only 20 or 32 byte addresses),
	// which is good, but not required for these tests as they make code less reuable
	// appparams.SetAddressPrefixes()

	for _, tc := range []struct {
		desc             string
		denom            string
		expectedSubdenom string
		err              error
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "normal",
			denom:            "factory/" + tfDenomCreator + "/bitcoin",
			expectedSubdenom: "bitcoin",
		},
		{
			desc:             "multiple slashes in subdenom",
			denom:            "factory/" + tfDenomCreator + "/bitcoin/1",
			expectedSubdenom: "bitcoin/1",
		},
		{
			desc:             "no subdenom",
			denom:            "factory/" + tfDenomCreator + "/",
			expectedSubdenom: "",
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/" + tfDenomCreator + "/bitcoin",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "subdenom of only slashes",
			denom:            "factory/" + tfDenomCreator + "/////",
			expectedSubdenom: "////",
		},
		{
			desc:  "too long name",
			denom: "factory/" + tfDenomCreator + "/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			err:   types.ErrInvalidDenom,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			expectedCreator := tfDenomCreator
			creator, subdenom, err := types.DeconstructDenom(tc.denom)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedCreator, creator)
				require.Equal(t, tc.expectedSubdenom, subdenom)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	// appparams.SetAddressPrefixes()
	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  tfDenomCreator,
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  tfDenomCreator,
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  tfDenomCreator,
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  tfDenomCreator,
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  tfDenomCreator,
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  tfDenomCreator,
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			// GetTokenDenom does not parse creator as a bech32 address, it only
			// checks the length is <= MaxCreatorLength (75). Padding the
			// addr_safro test address with extra ASCII keeps the test focused
			// on the boundary check while remaining compatible with
			// sdk.ValidateDenom's character allow-list.
			desc:     "creator is exactly max length",
			creator:  tfDenomCreatorMax,
			subdenom: "bitcoin",
			valid:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
