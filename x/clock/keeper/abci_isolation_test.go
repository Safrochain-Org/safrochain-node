package keeper_test

import (
	"sort"
	"strings"

	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// highGasClockContract increments its counter and then burns its whole gas
// budget in clock_end_block. Copy of
// interchaintest/contracts/clock_example_high_gas.wasm.
//
//go:embed testdata/clock_example_high_gas.wasm
var highGasClockContract []byte

// instantiateAndRegister stores wasm, instantiates it and registers it with
// x/clock. The shared testutil helper always instantiates code ID 1, so it
// cannot be used to mix different contracts in one test.
func (s *KeeperTestSuite) instantiateAndRegister(wasm []byte) string {
	_, _, creator := testdata.KeyTestPubAddr()
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	ck := s.App.AppKeepers.ContractKeeper
	codeID, _, err := ck.Create(s.Ctx, creator, wasm, nil)
	s.Require().NoError(err)
	addr, _, err := ck.Instantiate(s.Ctx, codeID, creator, nil, []byte(`{}`), "clock", nil)
	s.Require().NoError(err)

	s.Require().NoError(s.App.AppKeepers.ClockKeeper.RegisterContract(s.Ctx, creator.String(), addr.String()))
	return addr.String()
}

func (s *KeeperTestSuite) isJailed(contract string) bool {
	c, err := s.App.AppKeepers.ClockKeeper.GetClockContract(s.Ctx, contract)
	s.Require().NoError(err)
	return c.IsJailed
}

// A failing clock contract must only jail itself. The EndBlocker used to reuse
// the error of the previous contract, so every contract after the first
// failure was jailed without being executed.
func (s *KeeperTestSuite) TestFailingContractDoesNotJailOthers() {
	s.SetupTest()

	var failing, healthy []string
	for range 4 {
		failing = append(failing, s.instantiateAndRegister(burnContract))
		healthy = append(healthy, s.instantiateAndRegister(clockContract))
	}

	// Precondition: at least one healthy contract is iterated after a failing
	// one (the store iterates in address order).
	all := append(append([]string{}, failing...), healthy...)
	sort.Strings(all)
	firstFailing := len(all)
	for i, c := range all {
		for _, f := range failing {
			if c == f && i < firstFailing {
				firstFailing = i
			}
		}
	}
	healthyAfterFailure := false
	for i, c := range all {
		for _, h := range healthy {
			if c == h && i > firstFailing {
				healthyAfterFailure = true
			}
		}
	}
	s.Require().True(healthyAfterFailure, "test precondition: a healthy contract must sort after a failing one")

	s.EndBlock()

	for _, f := range failing {
		s.Require().True(s.isJailed(f), "failing contract %s must be jailed", f)
	}
	for _, h := range healthy {
		s.Require().False(s.isJailed(h), "healthy contract %s must not be jailed", h)
		s.Require().Equal(int64(1), s.queryContract(h), "healthy contract %s must have run", h)
	}
}

// A contract whose clock_end_block fails must not keep the state changes it
// made before failing.
func (s *KeeperTestSuite) TestFailedExecutionIsReverted() {
	s.SetupTest()

	contract := s.instantiateAndRegister(highGasClockContract)
	s.Require().Equal(int64(0), s.queryContract(contract))

	s.EndBlock()

	s.Require().True(s.isJailed(contract))
	s.Require().Equal(int64(0), s.queryContract(contract), "state written before the failure must be reverted")
}

// The same contract cannot be registered twice through a differently-cased
// bech32 address, and malformed addresses return an error instead of
// panicking.
func (s *KeeperTestSuite) TestContractAddressIsNormalised() {
	s.SetupTest()
	k := s.App.AppKeepers.ClockKeeper

	contract := s.instantiateAndRegister(clockContract)
	info := s.App.AppKeepers.WasmKeeper.GetContractInfo(s.Ctx, sdk.MustAccAddressFromBech32(contract))

	err := k.RegisterContract(s.Ctx, info.Creator, strings.ToUpper(contract))
	s.Require().Error(err)

	contracts, err := k.GetAllContracts(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(contracts, 1)

	s.Require().NotPanics(func() {
		ok, err := k.IsContractManager(s.Ctx, info.Creator, "not-an-address")
		s.Require().False(ok)
		s.Require().Error(err)
	})
}
