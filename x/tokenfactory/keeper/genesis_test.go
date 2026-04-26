package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	s.SetupTestForInitGenesis()
	// Bech32 addresses use the chain's configured account prefix
	// (addr_safro). Using any other prefix or an invalid checksum makes
	// InitGenesis panic when the tokenfactory keeper round-trips each
	// denom's creator/admin through sdk.AccAddressFromBech32.
	const (
		creator   = "addr_safro1ksw2e59xqrkj8cju74fet6tz8etkjg29uflsey"
		diffAdmin = "addr_safro1m502pj7dp6csn6ufqfdvxfru8rgtlfmaun00t5"
	)
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/" + creator + "/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: creator,
				},
			},
			{
				Denom: "factory/" + creator + "/diff-admin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: diffAdmin,
				},
			},
			{
				Denom: "factory/" + creator + "/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: creator,
				},
			},
		},
	}

	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			s.App.AppKeepers.BankKeeper.SetDenomMetaData(s.Ctx, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}

	// check before initGenesis that the module account is nil
	tokenfactoryModuleAccount := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().Nil(tokenfactoryModuleAccount)

	err := s.App.AppKeepers.TokenFactoryKeeper.SetParams(s.Ctx, types.Params{DenomCreationFee: sdk.Coins{sdk.NewInt64Coin("usaf", 100)}})
	s.Require().NoError(err)
	s.App.AppKeepers.TokenFactoryKeeper.InitGenesis(s.Ctx, genesisState)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().NotNil(tokenfactoryModuleAccount)

	exportedGenesis := s.App.AppKeepers.TokenFactoryKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(exportedGenesis)
	s.Require().Equal(genesisState, *exportedGenesis)

	// verify that the exported bank genesis is valid
	err = s.App.AppKeepers.BankKeeper.SetParams(s.Ctx, banktypes.DefaultParams())
	s.Require().NoError(err)
	exportedBankGenesis := s.App.AppKeepers.BankKeeper.ExportGenesis(s.Ctx)
	s.Require().NoError(exportedBankGenesis.Validate())

	s.App.AppKeepers.BankKeeper.InitGenesis(s.Ctx, exportedBankGenesis)
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, check whether bank metadata is not replaced if i != 0, to cover both cases.
		if i != 0 {
			metadata, found := s.App.AppKeepers.BankKeeper.GetDenomMetaData(s.Ctx, denom.GetDenom())
			s.Require().True(found)
			s.Require().EqualValues(metadata, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}
}
