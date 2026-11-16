package keeper_test

import (
	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestSetParamsValidates() {
	k := s.App.AppKeepers.TokenFactoryKeeper

	err := k.SetParams(s.Ctx, types.Params{DenomCreationGasConsume: types.MaxDenomCreationGasConsume + 1})
	s.Require().Error(err)

	// stored params were not modified by the rejected update
	s.Require().Equal(types.DefaultParams().DenomCreationGasConsume, k.GetParams(s.Ctx).DenomCreationGasConsume)

	s.Require().NoError(k.SetParams(s.Ctx, types.DefaultParams()))
}

func (s *KeeperTestSuite) TestGetParamsDefaultsWhenAbsent() {
	k := s.App.AppKeepers.TokenFactoryKeeper

	store := s.App.AppKeepers.GetKey(types.StoreKey)
	s.Require().NotNil(store)
	kv := s.Ctx.KVStore(store)
	kv.Delete(types.ParamsKey)

	s.Require().Equal(types.DefaultParams(), k.GetParams(s.Ctx))
}
