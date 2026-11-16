package bindings_test

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	sdkmath "cosmossdk.io/math"

	bindings "github.com/Safrochain_Org/safrochain/wasmbindings"
	types "github.com/Safrochain_Org/safrochain/wasmbindings/types"
)

func (s *BindingsTestSuite) TestRejectAmbiguousCustomMsg() {
	actor := s.RandomAccountAddress()
	// wrapped messenger is nil: an ambiguous msg must be rejected before dispatch.
	messenger := bindings.CustomMessageDecorator(s.App.AppKeepers.BankKeeper, &s.App.AppKeepers.TokenFactoryKeeper)(nil)

	bz, err := json.Marshal(types.TokenFactoryMsg{
		CreateDenom: &types.CreateDenom{Subdenom: "MOON"},
		MintTokens:  &types.MintTokens{Denom: "x", Amount: sdkmath.NewInt(1), MintToAddress: actor.String()},
	})
	s.Require().NoError(err)

	_, _, _, err = messenger.DispatchMsg(s.Ctx, actor, "", wasmvmtypes.CosmosMsg{Custom: bz}) //nolint:dogsled // only the error matters here
	s.Require().ErrorContains(err, "exactly one variant")
}

func (s *BindingsTestSuite) TestRejectAmbiguousCustomQuery() {
	qp := bindings.NewQueryPlugin(s.App.AppKeepers.BankKeeper, &s.App.AppKeepers.TokenFactoryKeeper)
	querier := bindings.CustomQuerier(qp)

	bz, err := json.Marshal(types.TokenFactoryQuery{
		Params:    &types.GetParams{},
		FullDenom: &types.FullDenom{CreatorAddr: s.RandomBech32AccountAddress(), Subdenom: "MOON"},
	})
	s.Require().NoError(err)
	_, err = querier(s.Ctx, bz)
	s.Require().ErrorContains(err, "exactly one variant")

	// a single variant still works
	bz, err = json.Marshal(types.TokenFactoryQuery{Params: &types.GetParams{}})
	s.Require().NoError(err)
	_, err = querier(s.Ctx, bz)
	s.Require().NoError(err)
}
