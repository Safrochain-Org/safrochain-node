package keeper

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

type GovHooks struct {
	k Keeper
}

var _ govtypes.GovHooks = GovHooks{}

func (k Keeper) GovHooks() GovHooks {
	return GovHooks{k: k}
}

type Proposal struct {
	ProposalID uint64 `json:"proposal_id"`
	Proposer   string `json:"proposer"`
	Status     uint   `json:"status"`
	SubmitTime string `json:"submit_time"`
	Metadata   string `json:"metadata"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
}

func NewProposal(prop v1.Proposal) Proposal {
	return Proposal{
		ProposalID: prop.Id,
		Proposer:   prop.Proposer,
		Status:     uint(prop.Status),
		SubmitTime: formatSubmitTime(prop),
		Metadata:   prop.GetMetadata(),
		Title:      prop.GetTitle(),
		Summary:    prop.GetSummary(),
	}
}

// formatSubmitTime returns the proposal submission time in RFC 3339 (UTC), or
// an empty string if it is unset. It previously returned only the
// seconds-of-minute component of the time.
func formatSubmitTime(prop v1.Proposal) string {
	if prop.SubmitTime == nil {
		return ""
	}
	return prop.SubmitTime.UTC().Format(time.RFC3339Nano)
}

type Vote struct {
	ProposalID   uint64                   `json:"proposal_id"`
	VoterAddress string                   `json:"voter_address"`
	VoteOption   []*v1.WeightedVoteOption `json:"vote_option"`
}

func NewVote(vote v1.Vote) Vote {
	return Vote{
		ProposalID:   vote.ProposalId,
		VoterAddress: vote.Voter,
		VoteOption:   vote.Options,
	}
}

type SudoMsgAfterProposalSubmission struct {
	AfterProposalSubmission Proposal `json:"after_proposal_submission"`
}

type SudoMsgAfterProposalDeposit struct {
	AfterProposalDeposit Proposal `json:"after_proposal_deposit"`
}

type SudoMsgAfterProposalVote struct {
	AfterProposalVote Vote `json:"after_proposal_vote"`
}

type SudoAfterProposalVotingPeriodEnded struct {
	AfterProposalVotingPeriodEnded string `json:"after_proposal_voting_period_ended"`
}

// The GovHooks methods below never return an error: x/gov returns hook errors
// from proposal submission, deposits and votes, so a failing contract would
// censor governance — including any proposal meant to fix the problem. See
// ExecuteMessageOnContracts.

// dispatch encodes msg and delivers it to every registered governance
// contract.
func (h GovHooks) dispatch(ctx context.Context, hook string, msg any) {
	msgBz, err := json.Marshal(msg)
	if err != nil {
		h.k.Logger(ctx).Error("failed to encode governance hook message", "hook", hook, "error", err)
		return
	}

	h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixGov, msgBz)
}

// proposal loads the proposal a hook refers to. It reports false, after
// logging, when the proposal cannot be loaded.
func (h GovHooks) proposal(ctx context.Context, hook string, proposalID uint64) (v1.Proposal, bool) {
	prop, err := h.k.govKeeper.Proposals.Get(ctx, proposalID)
	if err != nil {
		h.k.Logger(ctx).Error("skipping governance hook: proposal not found", "hook", hook, "proposal_id", proposalID, "error", err)
		return v1.Proposal{}, false
	}

	return prop, true
}

func (h GovHooks) AfterProposalSubmission(ctx context.Context, proposalID uint64) error {
	const hook = "AfterProposalSubmission"
	if prop, ok := h.proposal(ctx, hook, proposalID); ok {
		h.dispatch(ctx, hook, SudoMsgAfterProposalSubmission{AfterProposalSubmission: NewProposal(prop)})
	}

	return nil
}

func (h GovHooks) AfterProposalDeposit(ctx context.Context, proposalID uint64, _ sdk.AccAddress) error {
	const hook = "AfterProposalDeposit"
	if prop, ok := h.proposal(ctx, hook, proposalID); ok {
		h.dispatch(ctx, hook, SudoMsgAfterProposalDeposit{AfterProposalDeposit: NewProposal(prop)})
	}

	return nil
}

func (h GovHooks) AfterProposalVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	const hook = "AfterProposalVote"
	vote, err := h.k.govKeeper.Votes.Get(ctx, collections.Join(proposalID, voterAddr))
	if err != nil {
		h.k.Logger(ctx).Error("skipping governance hook: vote not found", "hook", hook, "proposal_id", proposalID, "voter", voterAddr.String(), "error", err)
		return nil
	}

	h.dispatch(ctx, hook, SudoMsgAfterProposalVote{AfterProposalVote: NewVote(vote)})

	return nil
}

func (GovHooks) AfterProposalFailedMinDeposit(_ context.Context, _ uint64) error {
	return nil
}

func (h GovHooks) AfterProposalVotingPeriodEnded(ctx context.Context, proposalID uint64) error {
	h.dispatch(ctx, "AfterProposalVotingPeriodEnded", SudoAfterProposalVotingPeriodEnded{
		AfterProposalVotingPeriodEnded: strconv.FormatUint(proposalID, 10),
	})

	return nil
}
