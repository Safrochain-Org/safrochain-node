package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func (s *KeeperTestSuite) setNativeMetadata() {
	s.App.AppKeepers.BankKeeper.SetDenomMetaData(s.Ctx, banktypes.Metadata{
		Description: "native",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: "usaf", Exponent: 0},
			{Denom: "saf", Exponent: 6},
		},
		Base:    "usaf",
		Display: "saf",
		Name:    "Saf",
		Symbol:  "SAF",
	})
}

func (s *KeeperTestSuite) TestSetDenomMetadataSpoofGuard() {
	s.setNativeMetadata()
	s.CreateDefaultDenom()
	base := s.defaultDenom
	admin := s.TestAccs[0].String()

	unit := func(d string, e uint32) *banktypes.DenomUnit { return &banktypes.DenomUnit{Denom: d, Exponent: e} }

	specs := map[string]struct {
		md     banktypes.Metadata
		expErr bool
	}{
		"benign metadata": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{unit(base, 0), unit("bitc", 6)},
				Base:       base, Display: "bitc", Name: "Bitcoin", Symbol: "BTC",
			},
		},
		"symbol impersonates native": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{unit(base, 0)},
				Base:       base, Display: base, Name: "Bitcoin", Symbol: "saf",
			},
			expErr: true,
		},
		"name impersonates native (case-insensitive)": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{unit(base, 0)},
				Base:       base, Display: base, Name: "SAF", Symbol: "BTC",
			},
			expErr: true,
		},
		"display impersonates native via extra unit": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{unit(base, 0), unit("saf", 6)},
				Base:       base, Display: "saf", Name: "Bitcoin", Symbol: "BTC",
			},
			expErr: true,
		},
		"extra unit equals bond denom": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{unit(base, 0), unit("usaf", 6)},
				Base:       base, Display: base, Name: "Bitcoin", Symbol: "BTC",
			},
			expErr: true,
		},
		"extra unit is an ibc denom": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{unit(base, 0), unit("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", 6)},
				Base:       base, Display: base, Name: "Bitcoin", Symbol: "BTC",
			},
			expErr: true,
		},
		"extra unit has existing supply": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{unit(base, 0), unit("usecond", 6)},
				Base:       base, Display: base, Name: "Bitcoin", Symbol: "BTC",
			},
			expErr: true,
		},
		"alias equals bond denom": {
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{Denom: base, Exponent: 0, Aliases: []string{"usaf"}}},
				Base:       base, Display: base, Name: "Bitcoin", Symbol: "BTC",
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		s.Run(name, func() {
			_, err := s.msgServer.SetDenomMetadata(s.Ctx, &types.MsgSetDenomMetadata{Sender: admin, Metadata: spec.md})
			if spec.expErr {
				s.Require().ErrorIs(err, types.ErrInvalidMetadata)
				return
			}
			s.Require().NoError(err)
		})
	}
}
