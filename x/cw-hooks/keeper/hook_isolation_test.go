package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// failingContract exports no `sudo` entry point, so every hook delivered to it
// fails. See testdata/README.md.
//
//go:embed testdata/no_sudo.wasm
var failingContract []byte

// hookHeight is past the keeper's gentx skip height, so hooks are delivered.
const hookHeight = 10

// storeAndInstantiate stores wasm and instantiates the resulting code (the
// shared testutil helper always instantiates code ID 1).
func (s *KeeperTestSuite) storeAndInstantiate(creator sdk.AccAddress, wasm []byte) sdk.AccAddress {
	ck := s.App.AppKeepers.ContractKeeper
	codeID, _, err := ck.Create(s.Ctx, creator, wasm, nil)
	s.Require().NoError(err)

	addr, _, err := ck.Instantiate(s.Ctx, codeID, creator, nil, []byte(`{}`), "hook", nil)
	s.Require().NoError(err)

	return addr
}

func (s *KeeperTestSuite) bondDenom() string {
	denom, err := s.App.AppKeepers.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	return denom
}

// registerFailingHook registers a contract that fails on every hook delivery.
func (s *KeeperTestSuite) registerFailingHook(staking, gov bool) sdk.AccAddress {
	_, _, creator := testdata.KeyTestPubAddr()
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin(s.bondDenom(), sdkmath.NewInt(1_000_000))))

	contract := s.storeAndInstantiate(creator, failingContract)
	s.registerHook(contract, staking, gov)

	return contract
}

// authority is the governance authority, the only account that can register
// hook contracts.
func (s *KeeperTestSuite) authority() string {
	return s.App.AppKeepers.CWHooksKeeper.GetAuthority()
}

// registerHook registers contract for the requested hook types through the
// governance authority.
func (s *KeeperTestSuite) registerHook(contract sdk.AccAddress, staking, gov bool) {
	if staking {
		_, err := s.msgServer.RegisterStaking(s.Ctx, &types.MsgRegisterStaking{
			ContractAddress: contract.String(),
			RegisterAddress: s.authority(),
		})
		s.Require().NoError(err)
	}

	if gov {
		_, err := s.msgServer.RegisterGovernance(s.Ctx, &types.MsgRegisterGovernance{
			ContractAddress: contract.String(),
			RegisterAddress: s.authority(),
		})
		s.Require().NoError(err)
	}
}

// requireHookFailureEvent asserts that the isolated failure of contract was
// reported through an event instead of being returned to the caller.
func (s *KeeperTestSuite) requireHookFailureEvent(ctx sdk.Context, contract sdk.AccAddress) {
	for _, ev := range ctx.EventManager().Events() {
		if ev.Type != types.EventTypeContractExecutionFailed {
			continue
		}
		for _, attr := range ev.Attributes {
			if attr.Key == types.AttributeKeyContract && attr.Value == contract.String() {
				return
			}
		}
	}
	s.Failf("missing hook failure event", "no %s event for contract %s", types.EventTypeContractExecutionFailed, contract)
}

// A failing staking hook must not make delegations fail.
func (s *KeeperTestSuite) TestFailingStakingHookDoesNotBlockDelegations() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	validators, err := s.stakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(validators)

	contract := s.registerFailingHook(true, false)

	_, _, delegator := testdata.KeyTestPubAddr()
	amount := sdk.NewCoin(s.bondDenom(), sdkmath.NewInt(1_000_000))
	s.FundAcc(delegator, sdk.NewCoins(amount.AddAmount(amount.Amount)))

	stakingMsgServer := stakingkeeper.NewMsgServerImpl(&s.stakingKeeper)
	msg := &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validators[0].GetOperator(),
		Amount:           amount,
	}

	// First delegation (AfterDelegationModified hook).
	_, err = stakingMsgServer.Delegate(s.Ctx, msg)
	s.Require().NoError(err)

	// Second delegation (BeforeDelegationSharesModified hook).
	_, err = stakingMsgServer.Delegate(s.Ctx, msg)
	s.Require().NoError(err)

	s.requireHookFailureEvent(s.Ctx, contract)
}

// A failing staking hook must not abort the x/staking EndBlocker. In SDK v0.50
// AfterValidatorBonded errors are returned from the EndBlocker, which halts
// the chain.
func (s *KeeperTestSuite) TestFailingStakingHookDoesNotHaltEndBlocker() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	// Created before the hook is registered; it becomes bonded at the next
	// validator-set update, which fires AfterValidatorBonded.
	valAddr := s.SetupValidator(stakingtypes.Unbonded)

	contract := s.registerFailingHook(true, false)

	_, err := s.App.AppKeepers.StakingKeeper.EndBlocker(s.Ctx)
	s.Require().NoError(err)

	val, err := s.stakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)
	s.Require().True(val.IsBonded())

	s.requireHookFailureEvent(s.Ctx, contract)
}

// A failing governance hook must not block proposal submission, deposits or
// votes.
func (s *KeeperTestSuite) TestFailingGovHookDoesNotBlockGovernance() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	contract := s.registerFailingHook(false, true)

	govKeeper := s.App.AppKeepers.GovKeeper
	params, err := govKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)

	_, _, proposer := testdata.KeyTestPubAddr()
	s.FundAcc(proposer, params.MinDeposit)

	proposal, err := govKeeper.SubmitProposal(s.Ctx, []sdk.Msg{}, "", "title", "summary", proposer, false)
	s.Require().NoError(err)

	votingStarted, err := govKeeper.AddDeposit(s.Ctx, proposal.Id, proposer, params.MinDeposit)
	s.Require().NoError(err)
	s.Require().True(votingStarted)

	err = govKeeper.AddVote(s.Ctx, proposal.Id, proposer, govv1.NewNonSplitVoteOption(govv1.OptionYes), "")
	s.Require().NoError(err)

	s.requireHookFailureEvent(s.Ctx, contract)
}

// A hook contract that fails must not prevent the remaining registered
// contracts from receiving the hook.
func (s *KeeperTestSuite) TestFailingHookDoesNotStarveOtherContracts() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	first := s.registerFailingHook(true, false)
	second := s.registerFailingHook(true, false)

	valAddr := s.SetupValidator(stakingtypes.Bonded)
	_, _, delegator := testdata.KeyTestPubAddr()
	amount := sdk.NewCoin(s.bondDenom(), sdkmath.NewInt(1_000_000))
	s.FundAcc(delegator, sdk.NewCoins(amount))

	_, err := stakingkeeper.NewMsgServerImpl(&s.stakingKeeper).Delegate(s.Ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: valAddr.String(),
		Amount:           amount,
	})
	s.Require().NoError(err)

	// Both contracts were attempted: each failure is reported on its own.
	s.requireHookFailureEvent(s.Ctx, first)
	s.requireHookFailureEvent(s.Ctx, second)
}
