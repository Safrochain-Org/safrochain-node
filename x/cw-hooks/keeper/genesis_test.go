package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// Missing params fall back to the defaults instead of a zero gas limit.
func (s *KeeperTestSuite) TestGetParamsFallsBackToDefaults() {
	s.SetupTest()
	k := s.App.AppKeepers.CWHooksKeeper

	store := s.App.AppKeepers.GetKey(types.StoreKey)
	s.Ctx.KVStore(store).Delete(types.ParamsKey)

	s.Require().Equal(types.DefaultParams(), k.GetParams(s.Ctx))
}

// SetParams rejects out-of-range gas limits (also through MsgUpdateParams).
func (s *KeeperTestSuite) TestSetParamsValidates() {
	s.SetupTest()
	k := s.App.AppKeepers.CWHooksKeeper

	s.Require().Error(k.SetParams(s.Ctx, types.NewParams(0)))
	s.Require().Error(k.SetParams(s.Ctx, types.NewParams(types.MaxContractGasLimit+1)))
	s.Require().NoError(k.SetParams(s.Ctx, types.NewParams(types.MaxContractGasLimit)))

	_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
		Authority: s.authority(),
		Params:    types.NewParams(0),
	})
	s.Require().Error(err)
}

// Genesis registrations must point to existing contracts, and round-trip.
func (s *KeeperTestSuite) TestInitGenesisContracts() {
	s.SetupTest()
	k := s.App.AppKeepers.CWHooksKeeper

	_, _, notAContract := testdata.KeyTestPubAddr()
	s.Require().Panics(func() {
		k.InitGenesis(s.Ctx, types.GenesisState{
			Params:                   types.DefaultParams(),
			StakingContractAddresses: []string{notAContract.String()},
		})
	})

	_, _, creator := testdata.KeyTestPubAddr()
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin(s.bondDenom(), sdkmath.NewInt(1_000_000))))
	contract := s.storeAndInstantiate(creator, failingContract)

	genesis := types.GenesisState{
		Params:                   types.DefaultParams(),
		StakingContractAddresses: []string{contract.String()},
		GovContractAddresses:     []string{contract.String()},
	}
	k.InitGenesis(s.Ctx, genesis)
	s.Require().Equal(&genesis, k.ExportGenesis(s.Ctx))
}
