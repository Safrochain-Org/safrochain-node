package keeper_test

import (
	"fmt"

	"github.com/Safrochain_Org/safrochain/x/drip/types"
)

func (s *KeeperTestSuite) TestDripInitGenesis() {
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
			// SAF-06: enabling Drip with an empty allow-list must panic on
			// genesis validation. Previously this configuration was treated
			// as a benign "module is on, but no one can use it yet" state,
			// which silently turned every later allow-list addition into an
			// immediate distribution authority.
			"custom genesis - drip enabled, no one allowed",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string(nil),
				},
			},
			true,
		},
		{
			"custom genesis - drip enabled, only one addr allowed",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string{"addr_safro1vc2894vzx0j74yqvg6yvt23stmeyx6pa6xfjf0"},
				},
			},
			false,
		},
		{
			"custom genesis - drip enabled, 2 addr allowed",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string{
						"addr_safro1vc2894vzx0j74yqvg6yvt23stmeyx6pa6xfjf0",
						"addr_safro15ew2xgxp3xc7esguq5yz4ymekru8357xvw73k0",
					},
				},
			},
			false,
		},
		{
			"custom genesis - drip enabled, address invalid",
			types.GenesisState{
				Params: types.Params{
					EnableDrip:       true,
					AllowedAddresses: []string{"addr_safro1v6vllollollollollolloldmljapdev4s827ql"},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			if tc.expPanic {
				s.Require().Panics(func() {
					s.App.AppKeepers.DripKeeper.InitGenesis(s.Ctx, tc.genesis)
				})
			} else {
				s.Require().NotPanics(func() {
					s.App.AppKeepers.DripKeeper.InitGenesis(s.Ctx, tc.genesis)
				})

				params := s.App.AppKeepers.DripKeeper.GetParams(s.Ctx)
				s.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}
