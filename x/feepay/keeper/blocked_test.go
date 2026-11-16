package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Safrochain_Org/safrochain/x/feepay/types"
)

// The FeePay module account must not be able to receive arbitrary transfers:
// funds sent there would be stranded and untracked by recorded contract
// balances. Funding goes through MsgFundFeePayContract, which uses
// SendCoinsFromAccountToModule and does not consult the blocklist.
func (s *KeeperTestSuite) TestFeePayModuleAccountIsBlocked() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(
		sdk.NewCoin("stake", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("usaf", sdkmath.NewInt(100_000_000)),
	))

	feepayAddr := authtypes.NewModuleAddress(types.ModuleName)
	s.Require().True(s.bankKeeper.BlockedAddr(feepayAddr))

	// A plain MsgSend to the module account is rejected.
	bankMsgServer := bankkeeper.NewMsgServerImpl(s.bankKeeper)
	_, err := bankMsgServer.Send(s.Ctx, &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   feepayAddr.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("usaf", sdkmath.NewInt(1_000))),
	})
	s.Require().Error(err)

	// Funding a registered contract still works.
	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	s.registerFeePayContract(sender.String(), contract, 0, 1)

	_, err = s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contract,
		Amount:          sdk.NewCoins(sdk.NewCoin("usaf", sdkmath.NewInt(1_000_000))),
	})
	s.Require().NoError(err)

	fpc, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1_000_000), fpc.Balance)
	s.Require().Equal(sdkmath.NewInt(1_000_000), s.bankKeeper.GetBalance(s.Ctx, feepayAddr, "usaf").Amount)
}
