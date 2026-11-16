package keeper_test

import (
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	globalerrors "github.com/Safrochain_Org/safrochain/app/helpers"
	"github.com/Safrochain_Org/safrochain/x/feepay/types"
)

func (s *KeeperTestSuite) TestUnregisterClearsUsageEntries() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	_, _, wallet := testdata.KeyTestPubAddr()

	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	s.registerFeePayContract(sender.String(), contract, 0, 10)
	k := s.App.AppKeepers.FeePayKeeper

	fpc, err := k.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	s.Require().NoError(k.IncrementContractUses(s.Ctx, fpc, wallet.String(), 3))
	uses, err := k.GetContractUses(s.Ctx, fpc, wallet.String())
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), uses)

	_, err = s.msgServer.UnregisterFeePayContract(s.Ctx, &types.MsgUnregisterFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contract,
	})
	s.Require().NoError(err)

	s.registerFeePayContract(sender.String(), contract, 0, 10)
	fpc, err = k.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	uses, err = k.GetContractUses(s.Ctx, fpc, wallet.String())
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), uses, "usage must be cleared on unregister")
}

func (s *KeeperTestSuite) TestContractAddressIsNormalised() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	upper := strings.ToUpper(contract)
	s.Require().NotEqual(contract, upper)

	s.registerFeePayContract(sender.String(), contract, 0, 1)

	// The same contract in upper-case bech32 is a duplicate, not a new record.
	_, err := s.msgServer.RegisterFeePayContract(s.Ctx, &types.MsgRegisterFeePayContract{
		SenderAddress:  sender.String(),
		FeePayContract: &types.FeePayContract{ContractAddress: upper, WalletLimit: 1},
	})
	s.Require().ErrorIs(err, globalerrors.ErrContractAlreadyRegistered)

	// Lookups with either spelling map to the same record.
	k := s.App.AppKeepers.FeePayKeeper
	s.Require().True(k.IsContractRegistered(s.Ctx, upper))
	fpc, err := k.GetContract(s.Ctx, upper)
	s.Require().NoError(err)
	s.Require().Equal(contract, fpc.ContractAddress)
	s.Require().Len(k.GetAllContracts(s.Ctx), 1)

	// Unregistering through the upper-case spelling removes the single record.
	_, err = s.msgServer.UnregisterFeePayContract(s.Ctx, &types.MsgUnregisterFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: upper,
	})
	s.Require().NoError(err)
	s.Require().False(k.IsContractRegistered(s.Ctx, contract))
}
