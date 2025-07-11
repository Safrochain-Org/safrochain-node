package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Safrochain_Org/safrochain/x/feeshare/types"
)

func (s *KeeperTestSuite) TestFeeShareInitGenesis() {
	testCases := []struct {
		name     string
		genesis  types.GenesisState
		expPanic bool
	}{
		{
			"default genesis",
			s.genesis,
			false,
		},
		{
			"custom genesis - feeshare disabled",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  false,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(50, 2),
					AllowedDenoms:   []string{"usaf"},
				},
			},
			false,
		},
		{
			"custom genesis - feeshare enabled, 0% developer shares",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  true,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(0, 2),
					AllowedDenoms:   []string{"usaf"},
				},
			},
			false,
		},
		{
			"custom genesis - feeshare enabled, 100% developer shares",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  true,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(100, 2),
					AllowedDenoms:   []string{"usaf"},
				},
			},
			false,
		},
		{
			"custom genesis - feeshare enabled, all denoms allowed",
			types.GenesisState{
				Params: types.Params{
					EnableFeeShare:  true,
					DeveloperShares: sdkmath.LegacyNewDecWithPrec(10, 2),
					AllowedDenoms:   []string(nil),
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.Reset()

			if tc.expPanic {
				s.Require().Panics(func() {
					s.App.AppKeepers.FeeShareKeeper.InitGenesis(s.Ctx, tc.genesis)
				})
			} else {
				s.Require().NotPanics(func() {
					s.App.AppKeepers.FeeShareKeeper.InitGenesis(s.Ctx, tc.genesis)
				})

				params := s.App.AppKeepers.FeeShareKeeper.GetParams(s.Ctx)
				s.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}
