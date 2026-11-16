package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/clock/types"
)

// instantiateManyAndRegister stores wasm once and registers n instances.
func (s *KeeperTestSuite) instantiateManyAndRegister(wasm []byte, n int) []string {
	_, _, creator := testdata.KeyTestPubAddr()
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	ck := s.App.AppKeepers.ContractKeeper
	codeID, _, err := ck.Create(s.Ctx, creator, wasm, nil)
	s.Require().NoError(err)

	addrs := make([]string, 0, n)
	for range n {
		addr, _, err := ck.Instantiate(s.Ctx, codeID, creator, nil, []byte(`{}`), "clock", nil)
		s.Require().NoError(err)
		s.Require().NoError(s.App.AppKeepers.ClockKeeper.RegisterContract(s.Ctx, creator.String(), addr.String()))
		addrs = append(addrs, addr.String())
	}

	return addrs
}

func (s *KeeperTestSuite) countJailed(addrs []string) int {
	n := 0
	for _, a := range addrs {
		if s.isJailed(a) {
			n++
		}
	}
	return n
}

// The gas consumed by clock contracts in one block is bounded, and contracts
// that do not fit are served in the following blocks.
func (s *KeeperTestSuite) TestEndBlockerGasBudget() {
	s.SetupTest()
	s.updateGasLimit(types.MinContractGasLimit)

	perBlock := int(types.MaxBlockGas / types.MinContractGasLimit)
	contracts := s.instantiateManyAndRegister(highGasClockContract, perBlock+perBlock/2)

	// Each contract burns its whole budget, so exactly perBlock fit.
	s.EndBlock()
	s.Require().Equal(perBlock, s.countJailed(contracts))

	// The rest run in the next block.
	s.EndBlock()
	s.Require().Equal(len(contracts), s.countJailed(contracts))
}
