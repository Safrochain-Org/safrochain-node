package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/Safrochain_Org/safrochain/x/feeshare/types"
)

// A bech32 address that is not an instantiated contract must be rejected with
// an error, not nil-dereference inside the msg handlers.
func (s *KeeperTestSuite) TestMsgsRejectNonContractAddress() {
	s.SetupTest()
	_, _, deployer := testdata.KeyTestPubAddr()
	_, _, withdrawer := testdata.KeyTestPubAddr()
	_, _, notAContract := testdata.KeyTestPubAddr()

	s.Require().NotPanics(func() {
		_, err := s.msgServer.RegisterFeeShare(s.Ctx, &types.MsgRegisterFeeShare{
			ContractAddress:   notAContract.String(),
			DeployerAddress:   deployer.String(),
			WithdrawerAddress: withdrawer.String(),
		})
		s.Require().ErrorIs(err, types.ErrFeeShareNoContractDeployed)
	})

	// Stale registrations (e.g. imported through genesis) for a non-contract
	// must not panic the update / cancel handlers either.
	s.App.AppKeepers.FeeShareKeeper.SetFeeShare(s.Ctx, types.FeeShare{
		ContractAddress:   notAContract.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: withdrawer.String(),
	})

	s.Require().NotPanics(func() {
		_, err := s.msgServer.UpdateFeeShare(s.Ctx, &types.MsgUpdateFeeShare{
			ContractAddress:   notAContract.String(),
			DeployerAddress:   deployer.String(),
			WithdrawerAddress: deployer.String(),
		})
		s.Require().ErrorIs(err, types.ErrFeeShareNoContractDeployed)
	})

	s.Require().NotPanics(func() {
		_, err := s.msgServer.CancelFeeShare(s.Ctx, &types.MsgCancelFeeShare{
			ContractAddress: notAContract.String(),
			DeployerAddress: deployer.String(),
		})
		s.Require().ErrorIs(err, types.ErrFeeShareNoContractDeployed)
	})
}
