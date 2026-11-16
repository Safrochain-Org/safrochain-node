package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/drip/types"
)

func (s *KeeperTestSuite) TestDripDistributeTokensMsgs() {
	bondDenom, err := s.App.AppKeepers.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	factoryDenom := "factory/" + sdk.AccAddress([]byte("tokenfactory_creator")).String() + "/sub"

	_, _, allowedSender := testdata.KeyTestPubAddr()
	_, _, notAllowedSender := testdata.KeyTestPubAddr()
	s.FundAcc(allowedSender, sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1_000_000))))
	s.FundAcc(notAllowedSender, sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1_000_000))))
	s.FundAcc(allowedSender, sdk.NewCoins(
		sdk.NewCoin(factoryDenom, sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", sdkmath.NewInt(1_000_000)),
	))

	_ = s.App.AppKeepers.DripKeeper.SetParams(s.Ctx, types.Params{
		EnableDrip: true,
		AllowedAddresses: []string{
			allowedSender.String(),
		},
	})

	for _, tc := range []struct {
		desc       string
		senderAddr string
		coins      sdk.Coins
		success    bool
	}{
		{
			desc:       "Success - Allowed sender with proper funds",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1))),
			success:    true,
		},
		{
			desc:       "Success - Allowed sender with tokenfactory denom",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin(factoryDenom, sdkmath.NewInt(1))),
			success:    true,
		},
		{
			desc:       "Success - Allowed sender with bond and tokenfactory denoms",
			senderAddr: allowedSender.String(),
			coins: sdk.NewCoins(
				sdk.NewCoin(bondDenom, sdkmath.NewInt(1)),
				sdk.NewCoin(factoryDenom, sdkmath.NewInt(1)),
			),
			success: true,
		},
		{
			desc:       "Fail - Allowed sender with funded but disallowed ibc denom",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", sdkmath.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Allowed sender with bond denom plus disallowed denom",
			senderAddr: allowedSender.String(),
			coins: sdk.NewCoins(
				sdk.NewCoin(bondDenom, sdkmath.NewInt(1)),
				sdk.NewCoin("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", sdkmath.NewInt(1)),
			),
			success: false,
		},
		{
			desc:       "Fail - Allowed sender no proper funds",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("notarealtoken", sdkmath.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Allowed sender no tokens",
			senderAddr: allowedSender.String(),
			coins:      nil,
			success:    false,
		},
		{
			desc:       "Fail - Allowed sender empty tokens",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(),
			success:    false,
		},
		{
			desc:       "Fail - No sender withproper funds",
			senderAddr: "",
			coins:      sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Non Allowed sender proper funds",
			senderAddr: notAllowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Non Allowed sender improper funds",
			senderAddr: notAllowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("notarealtoken", sdkmath.NewInt(1))),
			success:    false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			msg := types.MsgDistributeTokens{
				SenderAddress: tc.senderAddr,
				Amount:        tc.coins,
			}
			_, err := s.msgServer.DistributeTokens(s.Ctx, &msg)

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDripDistributeTokensDenomRestriction() {
	_, _, sender := testdata.KeyTestPubAddr()
	ibcDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	bondDenom, err := s.App.AppKeepers.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)

	s.FundAcc(sender, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdkmath.NewInt(10)),
		sdk.NewCoin(ibcDenom, sdkmath.NewInt(10)),
	))
	s.Require().NoError(s.App.AppKeepers.DripKeeper.SetParams(s.Ctx, types.Params{
		EnableDrip:       true,
		AllowedAddresses: []string{sender.String()},
	}))

	_, err = s.msgServer.DistributeTokens(s.Ctx, &types.MsgDistributeTokens{
		SenderAddress: sender.String(),
		Amount:        sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1))),
	})
	s.Require().ErrorIs(err, types.ErrDenomNotAllowed)

	_, err = s.msgServer.DistributeTokens(s.Ctx, &types.MsgDistributeTokens{
		SenderAddress: sender.String(),
		Amount:        sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1)), sdk.NewCoin(ibcDenom, sdkmath.NewInt(1))),
	})
	s.Require().ErrorIs(err, types.ErrDenomNotAllowed)

	_, err = s.msgServer.DistributeTokens(s.Ctx, &types.MsgDistributeTokens{
		SenderAddress: sender.String(),
		Amount:        sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1))),
	})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestUpdateDripParams() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	for _, tc := range []struct {
		desc             string
		isEnabled        bool
		AllowedAddresses []string
		success          bool
	}{
		{
			// SAF-06: enabling drip with an empty allow-list is no longer
			// accepted; operators must seed AllowedAddresses first.
			desc:             "Fail - Enabled with empty allow-list (SAF-06)",
			isEnabled:        true,
			AllowedAddresses: []string{},
			success:          false,
		},
		{
			desc:             "Success - Valid off",
			isEnabled:        false,
			AllowedAddresses: []string{},
			success:          true,
		},
		{
			desc:             "Success - On and 1 allowed address",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String()},
			success:          true,
		},
		{
			desc:             "Fail - On and 2 duplicate addresses",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String(), addr.String()},
			success:          false,
		},
		{
			desc:             "Success - On and 2 unique",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String(), addr2.String()},
			success:          true,
		},
		{
			desc:             "Success - On and 2 duplicate 1 unique",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String(), addr2.String(), addr.String()},
			success:          false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			params := types.Params{
				EnableDrip:       tc.isEnabled,
				AllowedAddresses: tc.AllowedAddresses,
			}
			err := s.App.AppKeepers.DripKeeper.SetParams(s.Ctx, params)

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
