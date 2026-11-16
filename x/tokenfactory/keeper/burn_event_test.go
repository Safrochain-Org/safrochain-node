package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestBurnEventAttributesBurnFromAddress() {
	s.CreateDefaultDenom()
	admin := s.TestAccs[0].String()
	holder := s.TestAccs[1].String()

	_, err := s.msgServer.Mint(s.Ctx, &types.MsgMint{
		Sender:        admin,
		Amount:        sdk.NewInt64Coin(s.defaultDenom, 10),
		MintToAddress: holder,
	})
	s.Require().NoError(err)

	burnEventFrom := func(msg *types.MsgBurn) string {
		ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
		_, err := s.msgServer.Burn(ctx, msg)
		s.Require().NoError(err)
		for _, ev := range ctx.EventManager().Events() {
			if ev.Type != "tf_burn" {
				continue
			}
			for _, attr := range ev.Attributes {
				if attr.Key == types.AttributeBurnFromAddress {
					return attr.Value
				}
			}
		}
		s.FailNow("tf_burn event not found")
		return ""
	}

	// burning from another account reports that account, not the sender
	s.Require().Equal(holder, burnEventFrom(&types.MsgBurn{
		Sender:          admin,
		Amount:          sdk.NewInt64Coin(s.defaultDenom, 5),
		BurnFromAddress: holder,
	}))

	// empty BurnFromAddress defaults to the sender
	_, err = s.msgServer.Mint(s.Ctx, &types.MsgMint{
		Sender:        admin,
		Amount:        sdk.NewInt64Coin(s.defaultDenom, 10),
		MintToAddress: admin,
	})
	s.Require().NoError(err)
	s.Require().Equal(admin, burnEventFrom(&types.MsgBurn{
		Sender: admin,
		Amount: sdk.NewInt64Coin(s.defaultDenom, 5),
	}))
}
