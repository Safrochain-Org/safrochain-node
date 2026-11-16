package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// registerBoth registers contractAddress for staking and governance hooks on
// behalf of registrar and returns the two errors.
func (s *KeeperTestSuite) registerBoth(contractAddress, registrar string) (error, error) {
	_, stakingErr := s.msgServer.RegisterStaking(s.Ctx, &types.MsgRegisterStaking{
		ContractAddress: contractAddress,
		RegisterAddress: registrar,
	})
	_, govErr := s.msgServer.RegisterGovernance(s.Ctx, &types.MsgRegisterGovernance{
		ContractAddress: contractAddress,
		RegisterAddress: registrar,
	})
	return stakingErr, govErr
}

// unregisterBoth unregisters contractAddress from staking and governance hooks
// on behalf of registrar and returns the two errors.
func (s *KeeperTestSuite) unregisterBoth(contractAddress, registrar string) (error, error) {
	_, stakingErr := s.msgServer.UnregisterStaking(s.Ctx, &types.MsgUnregisterStaking{
		ContractAddress: contractAddress,
		RegisterAddress: registrar,
	})
	_, govErr := s.msgServer.UnregisterGovernance(s.Ctx, &types.MsgUnregisterGovernance{
		ContractAddress: contractAddress,
		RegisterAddress: registrar,
	})
	return stakingErr, govErr
}

// Only the governance authority can register hook contracts.
func (s *KeeperTestSuite) TestRegisterContracts() {
	s.SetupTest()
	_, _, creator := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	// testutil.InstantiateContract attaches 1stake of funds.
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contract := s.InstantiateContract(creator.String(), "", wasmContract)
	contractWithAdmin := s.InstantiateContract(creator.String(), admin.String(), wasmContract)
	_, _, notAContract := testdata.KeyTestPubAddr()

	for _, tc := range []struct {
		desc      string
		contract  string
		registrar string
		shouldErr bool
	}{
		{desc: "invalid contract address", contract: "invalid", registrar: s.authority(), shouldErr: true},
		{desc: "invalid registrar address", contract: contract, registrar: "invalid", shouldErr: true},
		{desc: "contract creator cannot register", contract: contract, registrar: creator.String(), shouldErr: true},
		{desc: "contract admin cannot register", contract: contractWithAdmin, registrar: admin.String(), shouldErr: true},
		{desc: "address without a contract", contract: notAContract.String(), registrar: s.authority(), shouldErr: true},
		{desc: "governance authority registers", contract: contract, registrar: s.authority(), shouldErr: false},
		{desc: "already registered", contract: contract, registrar: s.authority(), shouldErr: true},
		{desc: "governance authority registers a contract with an admin", contract: contractWithAdmin, registrar: s.authority(), shouldErr: false},
	} {
		s.Run(tc.desc, func() {
			stakingErr, govErr := s.registerBoth(tc.contract, tc.registrar)
			if tc.shouldErr {
				s.Require().Error(stakingErr)
				s.Require().Error(govErr)
				return
			}
			s.Require().NoError(stakingErr)
			s.Require().NoError(govErr)
		})
	}
}

// The contract manager (admin, or creator when there is no admin) and the
// governance authority can unregister a contract; nobody else can.
func (s *KeeperTestSuite) TestUnRegisterContracts() {
	s.SetupTest()
	_, _, creator := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	_, _, stranger := testdata.KeyTestPubAddr()
	// testutil.InstantiateContract attaches 1stake of funds.
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	// No admin: the creator manages it.
	noAdmin := sdk.MustAccAddressFromBech32(s.InstantiateContract(creator.String(), "", wasmContract))
	// Admin differs from the creator (e.g. a contract handed over to a DAO):
	// the admin manages it, the creator no longer does.
	withAdmin := sdk.MustAccAddressFromBech32(s.InstantiateContract(creator.String(), admin.String(), wasmContract))
	// Removed by governance.
	byGov := sdk.MustAccAddressFromBech32(s.InstantiateContract(creator.String(), admin.String(), wasmContract))

	for _, c := range []sdk.AccAddress{noAdmin, withAdmin, byGov} {
		s.registerHook(c, true, true)
	}

	for _, tc := range []struct {
		desc      string
		contract  string
		registrar string
		shouldErr bool
	}{
		{desc: "invalid contract address", contract: "invalid", registrar: creator.String(), shouldErr: true},
		{desc: "invalid registrar address", contract: noAdmin.String(), registrar: "invalid", shouldErr: true},
		{desc: "stranger cannot unregister", contract: noAdmin.String(), registrar: stranger.String(), shouldErr: true},
		{desc: "creator unregisters a contract without admin", contract: noAdmin.String(), registrar: creator.String(), shouldErr: false},
		{desc: "already unregistered", contract: noAdmin.String(), registrar: creator.String(), shouldErr: true},
		{desc: "creator cannot unregister a contract with a different admin", contract: withAdmin.String(), registrar: creator.String(), shouldErr: true},
		{desc: "admin unregisters", contract: withAdmin.String(), registrar: admin.String(), shouldErr: false},
		{desc: "governance authority unregisters any contract", contract: byGov.String(), registrar: s.authority(), shouldErr: false},
	} {
		s.Run(tc.desc, func() {
			stakingErr, govErr := s.unregisterBoth(tc.contract, tc.registrar)
			if tc.shouldErr {
				s.Require().Error(stakingErr)
				s.Require().Error(govErr)
				return
			}
			s.Require().NoError(stakingErr)
			s.Require().NoError(govErr)
		})
	}

	sc, err := s.queryClient.StakingContracts(s.Ctx, &types.QueryStakingContractsRequest{})
	s.Require().NoError(err)
	s.Require().Empty(sc.Contracts)

	gc, err := s.queryClient.GovernanceContracts(s.Ctx, &types.QueryGovernanceContractsRequest{})
	s.Require().NoError(err)
	s.Require().Empty(gc.Contracts)
}

// Registration is carried by a governance proposal: the messages are routable
// and signed by the governance module account.
func (s *KeeperTestSuite) TestRegistrationThroughGovernanceProposal() {
	s.SetupTest()
	_, _, creator := testdata.KeyTestPubAddr()
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	contract := s.InstantiateContract(creator.String(), "", wasmContract)

	params, err := s.App.AppKeepers.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	_, _, proposer := testdata.KeyTestPubAddr()
	s.FundAcc(proposer, params.MinDeposit)

	_, err = s.App.AppKeepers.GovKeeper.SubmitProposal(s.Ctx, []sdk.Msg{
		&types.MsgRegisterStaking{ContractAddress: contract, RegisterAddress: s.authority()},
		&types.MsgRegisterGovernance{ContractAddress: contract, RegisterAddress: s.authority()},
	}, "", "register hooks", "register a vetted hook contract", proposer, false)
	s.Require().NoError(err)
}
