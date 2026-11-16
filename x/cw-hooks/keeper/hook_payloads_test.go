package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/keeper"
)

// BeforeDelegationCreated and AfterValidatorRemoved used to look up state that
// does not exist yet / any more, so they never delivered anything. They are now
// built from the hook arguments.
func (s *KeeperTestSuite) TestHooksWithoutStateAreDelivered() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	contract := s.registerFailingHook(true, false)
	hooks := s.App.AppKeepers.CWHooksKeeper.StakingHooks()

	_, _, delegator := testdata.KeyTestPubAddr()
	_, _, removed := testdata.KeyTestPubAddr()

	ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(hooks.BeforeDelegationCreated(ctx, delegator, sdk.ValAddress(removed)))
	s.requireHookFailureEvent(ctx, contract)

	ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(hooks.AfterValidatorRemoved(ctx, sdk.ConsAddress(removed), sdk.ValAddress(removed)))
	s.requireHookFailureEvent(ctx, contract)
}

// The proposal submit time is sent as an RFC 3339 timestamp; it used to be the
// seconds-of-minute component only.
func (s *KeeperTestSuite) TestProposalSubmitTimeFormat() {
	submitted := time.Date(2026, 11, 15, 12, 34, 56, 0, time.UTC)

	s.Require().Equal("2026-11-15T12:34:56Z", keeper.NewProposal(v1.Proposal{Id: 1, SubmitTime: &submitted}).SubmitTime)
	s.Require().Empty(keeper.NewProposal(v1.Proposal{Id: 1}).SubmitTime)
}
