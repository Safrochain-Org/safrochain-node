package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

const genesisTestCreator = "addr_safro1ksw2e59xqrkj8cju74fet6tz8etkjg29uflsey"

func (s *KeeperTestSuite) TestInitGenesisKeepsExistingMetadata() {
	s.SetupTestForInitGenesis()

	denom := "factory/" + genesisTestCreator + "/bitcoin"
	custom := banktypes.Metadata{
		Description: "custom",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: denom, Exponent: 0},
			{Denom: "btc", Exponent: 8},
		},
		Base:    denom,
		Display: "btc",
		Name:    "Bitcoin",
		Symbol:  "BTC",
	}
	s.App.AppKeepers.BankKeeper.SetDenomMetaData(s.Ctx, custom)

	s.App.AppKeepers.TokenFactoryKeeper.InitGenesis(s.Ctx, types.GenesisState{
		Params: types.DefaultParams(),
		FactoryDenoms: []types.GenesisDenom{{
			Denom:             denom,
			AuthorityMetadata: types.DenomAuthorityMetadata{Admin: genesisTestCreator},
		}},
	})

	got, found := s.App.AppKeepers.BankKeeper.GetDenomMetaData(s.Ctx, denom)
	s.Require().True(found)
	s.Require().Equal(custom, got)
}

func (s *KeeperTestSuite) TestInitGenesisRejectsInvalidState() {
	s.SetupTestForInitGenesis()
	denom := "factory/" + genesisTestCreator + "/bitcoin"

	s.Require().Panics(func() {
		s.App.AppKeepers.TokenFactoryKeeper.InitGenesis(s.Ctx, types.GenesisState{
			Params: types.DefaultParams(),
			FactoryDenoms: []types.GenesisDenom{{
				Denom: denom,
			}},
		})
	}, "empty admin")

	s.Require().Panics(func() {
		gd := types.GenesisDenom{Denom: denom, AuthorityMetadata: types.DenomAuthorityMetadata{Admin: genesisTestCreator}}
		s.App.AppKeepers.TokenFactoryKeeper.InitGenesis(s.Ctx, types.GenesisState{
			Params:        types.DefaultParams(),
			FactoryDenoms: []types.GenesisDenom{gd, gd},
		})
	}, "duplicate denom")
}

func (s *KeeperTestSuite) TestMintToModuleAccountMintsNothing() {
	s.CreateDefaultDenom()
	moduleAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)

	_, err := s.msgServer.Mint(s.Ctx, &types.MsgMint{
		Sender:        s.TestAccs[0].String(),
		Amount:        sdk.NewInt64Coin(s.defaultDenom, 10),
		MintToAddress: moduleAddr.String(),
	})
	s.Require().ErrorIs(err, types.ErrModuleAccount)

	// no coins were minted before the failure
	s.Require().True(s.App.AppKeepers.BankKeeper.GetSupply(s.Ctx, s.defaultDenom).Amount.IsZero())
}
