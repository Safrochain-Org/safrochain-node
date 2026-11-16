package bindings_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	bindings "github.com/Safrochain_Org/safrochain/wasmbindings"
	"github.com/Safrochain_Org/safrochain/wasmbindings/types"
	tftypes "github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func (s *BindingsTestSuite) TestSetMetadataSpoofGuard() {
	actor := s.RandomAccountAddress()
	s.StoreReflectCode(actor)
	fee := tftypes.DefaultParams().DenomCreationFee[0]
	s.FundAcc(actor, sdk.NewCoins(sdk.NewCoin(fee.Denom, fee.Amount.MulRaw(100))))

	s.App.AppKeepers.BankKeeper.SetDenomMetaData(s.Ctx, banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "usaf"}, {Denom: "saf", Exponent: 6}},
		Base:       "usaf", Display: "saf", Name: "Saf", Symbol: "SAF",
	})

	tf := &s.App.AppKeepers.TokenFactoryKeeper
	bk := s.App.AppKeepers.BankKeeper
	_, err := bindings.PerformCreateDenom(s.Ctx, tf, bk, actor, &types.CreateDenom{Subdenom: "MOON"})
	s.Require().NoError(err)
	denom := "factory/" + actor.String() + "/MOON"

	spoof := types.Metadata{
		Base:       denom,
		Display:    "saf",
		Name:       "Saf",
		Symbol:     "SAF",
		DenomUnits: []types.DenomUnit{{Denom: denom}, {Denom: "saf", Exponent: 6}},
	}

	// direct set-metadata is rejected
	s.Require().Error(bindings.PerformSetMetadata(s.Ctx, tf, bk, actor, denom, spoof))

	// create-denom with spoofed metadata is rejected as well
	sunDenom := "factory/" + actor.String() + "/SUN"
	spoofSun := spoof
	spoofSun.Base = sunDenom
	spoofSun.DenomUnits = []types.DenomUnit{{Denom: sunDenom}, {Denom: "saf", Exponent: 6}}
	_, err = bindings.PerformCreateDenom(s.Ctx, tf, bk, actor, &types.CreateDenom{Subdenom: "SUN", Metadata: &spoofSun})
	s.Require().Error(err)

	// benign metadata is accepted
	ok := types.Metadata{
		Base: denom, Display: "moon", Name: "Moon", Symbol: "MOON",
		DenomUnits: []types.DenomUnit{{Denom: denom}, {Denom: "moon", Exponent: 6}},
	}
	s.Require().NoError(bindings.PerformSetMetadata(s.Ctx, tf, bk, actor, denom, ok))
}
