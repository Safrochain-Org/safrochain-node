package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/feepay/types"
)

// The bond denom must be resolved at runtime from x/staking, not fixed at
// construction time (the live testnet and mainnet use "usaf").
func (s *KeeperTestSuite) TestFundAndUnregisterUseRuntimeBondDenom() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	s.registerFeePayContract(sender.String(), contract, 0, 1)

	// Switch the staking bond denom at runtime.
	const newDenom = "utestbond"
	params, err := s.App.AppKeepers.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	params.BondDenom = newDenom
	s.Require().NoError(s.App.AppKeepers.StakingKeeper.SetParams(s.Ctx, params))

	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin(newDenom, sdkmath.NewInt(5_000_000)), sdk.NewCoin("usaf", sdkmath.NewInt(5_000_000))))

	// Old denom is rejected, new bond denom accepted.
	_, err = s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contract,
		Amount:          sdk.NewCoins(sdk.NewCoin("usaf", sdkmath.NewInt(1_000_000))),
	})
	s.Require().Error(err)

	_, err = s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contract,
		Amount:          sdk.NewCoins(sdk.NewCoin(newDenom, sdkmath.NewInt(1_000_000))),
	})
	s.Require().NoError(err)

	// Unregistering refunds in the runtime bond denom.
	before := s.bankKeeper.GetBalance(s.Ctx, sender, newDenom)
	_, err = s.msgServer.UnregisterFeePayContract(s.Ctx, &types.MsgUnregisterFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contract,
	})
	s.Require().NoError(err)
	after := s.bankKeeper.GetBalance(s.Ctx, sender, newDenom)
	s.Require().Equal(sdkmath.NewInt(1_000_000), after.Amount.Sub(before.Amount))
}
