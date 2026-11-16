package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestCreateDenomRejectsEmptySubdenom() {
	_, err := s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   s.TestAccs[0].String(),
		Subdenom: "",
	})
	s.Require().ErrorIs(err, types.ErrInvalidDenom)
}

func (s *KeeperTestSuite) TestCreateDenomRejectsExistingSupply() {
	// A denom that has supply but no bank metadata must not be claimable.
	denom, err := types.GetTokenDenom(s.TestAccs[0].String(), "ghost")
	s.Require().NoError(err)
	s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewInt64Coin(denom, 1)))

	_, found := s.App.AppKeepers.BankKeeper.GetDenomMetaData(s.Ctx, denom)
	s.Require().False(found)
	s.Require().True(s.App.AppKeepers.BankKeeper.HasSupply(s.Ctx, denom))

	_, err = s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   s.TestAccs[0].String(),
		Subdenom: "ghost",
	})
	s.Require().ErrorIs(err, types.ErrDenomExists)
}
