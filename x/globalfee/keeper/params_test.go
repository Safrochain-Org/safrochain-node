package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/testutil"
	"github.com/Safrochain_Org/safrochain/x/globalfee/keeper"
	"github.com/Safrochain_Org/safrochain/x/globalfee/types"
)

type ParamsTestSuite struct {
	testutil.KeeperTestHelper
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (s *ParamsTestSuite) TestSetParamsValidates() {
	s.Setup()
	k := s.App.AppKeepers.GlobalFeeKeeper

	dec := func(denom string, amt int64) sdk.DecCoin {
		return sdk.DecCoin{Denom: denom, Amount: sdkmath.LegacyNewDec(amt)}
	}

	for _, tc := range []struct {
		desc      string
		prices    sdk.DecCoins
		shouldErr bool
	}{
		{"valid sorted", sdk.DecCoins{dec("uatom", 1), dec("usaf", 2)}, false},
		{"empty", nil, false},
		{"unsorted", sdk.DecCoins{dec("usaf", 1), dec("uatom", 2)}, true},
		{"duplicate", sdk.DecCoins{dec("usaf", 1), dec("usaf", 2)}, true},
		{"negative", sdk.DecCoins{dec("usaf", -1)}, true},
		{"invalid denom", sdk.DecCoins{dec("1", 1)}, true},
	} {
		s.Run(tc.desc, func() {
			before := k.GetParams(s.Ctx)

			err := k.SetParams(s.Ctx, types.Params{MinimumGasPrices: tc.prices})
			if tc.shouldErr {
				s.Require().Error(err)
				// Invalid params must not be persisted.
				s.Require().Equal(before, k.GetParams(s.Ctx))
			} else {
				s.Require().NoError(err)
			}

			// The governance path goes through the same validation.
			_, err = keeper.NewMsgServerImpl(k).UpdateParams(s.Ctx, &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    types.Params{MinimumGasPrices: tc.prices},
			})
			if tc.shouldErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
