# Wrappers

This module provides a **governance wrapper** around the standard Cosmos SDK `x/gov` module. It allows the chain to customize, extend, and fix behaviors in the upstream governance module without modifying the SDK source directly.

## Abstract

The Wrappers module replicates the full governance API (proposals, voting, deposits, tallying, parameters) while adding Safrochain-specific enhancements and bug fixes. The upstream `cosmos-sdk/x/gov` module is embedded via composition — a `KeeperWrapper` struct wraps the standard `govkeeper.Keeper` and a custom gRPC query server overrides the default one.

## Concepts

- **KeeperWrapper** — wraps `govkeeper.Keeper` from the Cosmos SDK, forwarding all standard operations while allowing Safrochain to inject custom behavior (e.g. hooks, query fixes).
- **Custom Query Server** — replaces the default `x/gov` query server with a fixed implementation that addresses the pagination panic that occurs when querying proposals from removed modules.
- **Constitution query** — a Safrochain-specific gRPC endpoint (`Query/Constitution`) that returns the chain's constitution text, a feature not available in the upstream SDK at the time of forking.
- **Migration support** — the module registers all upstream migrations (v1→v2→v3→v4→v5) so the governance state stays compatible with SDK upgrades.

## State

The module stores all governance state through the underlying `cosmos-sdk/x/gov` keeper. Key stored objects include:

| Object | Key | Description |
|--------|-----|-------------|
| Proposals | `Proposals` (collections.Map[uint64, v1.Proposal]) | All governance proposals |
| Votes | `Votes` (collections.Map[Pair[uint64, sdk.AccAddress], v1.Vote]) | Votes cast on proposals |
| Deposits | `Deposits` (collections.Map[Pair[uint64, sdk.AccAddress], v1.Deposit]) | Deposits made on proposals |
| Params | `Params` (collections.Item[v1.Params]) | Governance parameters |
| Constitution | `Constitution` (collections.Item[string]) | Chain constitution text |

## Messages

| Message | Action |
|---------|--------|
| `MsgSubmitProposal` | Submit a new governance proposal |
| `MsgDeposit` | Deposit tokens to activate a proposal into voting period |
| `MsgVote` | Cast a vote on an active proposal |
| `MsgVoteWeighted` | Cast a weighted vote (split voting power across options) |
| `MsgExecLegacyContent` | Execute a legacy proposal's content |

Messages are handled by the standard `govkeeper.NewMsgServerImpl` from the Cosmos SDK — the wrapper delegates message handling to the embedded keeper.

## Queries

All standard `x/gov` gRPC queries are available:

| Query | Description |
|-------|-------------|
| `Query/Proposal` | Get a single proposal by ID |
| `Query/Proposals` | List proposals with pagination and filtering (by status, voter, depositor) |
| `Query/Vote` | Get a specific vote |
| `Query/Votes` | List all votes on a proposal |
| `Query/Deposit` | Get a specific deposit |
| `Query/Deposits` | List all deposits on a proposal |
| `Query/Params` | Get governance parameters (deposit, voting, tallying) |
| `Query/TallyResult` | Get the current tally result for a proposal |
| `Query/Constitution` | Get the chain's constitution text |

Legacy v1beta1 equivalents are also served for backward compatibility.

## Params

Module parameters are identical to upstream `x/gov`:

| Parameter | Type | Description |
|-----------|------|-------------|
| `min_deposit` | Array of Coin | Minimum deposit required for a proposal to enter voting period |
| `max_deposit_period` | Duration | Maximum time for depositing on a proposal |
| `voting_period` | Duration | Duration of the voting period |
| `quorum` | String (dec) | Minimum percentage of voting power that must vote |
| `threshold` | String (dec) | Minimum percentage of votes required to pass |
| `veto_threshold` | String (dec) | Percentage of votes needed to veto |
| `min_initial_deposit_ratio` | String (dec) | Ratio of min_deposit that must be paid by the submitter |

## Genesis

Governance genesis state mirrors the upstream `x/gov` genesis format:

- `proposals` — list of initial proposals
- `votes` — list of initial votes
- `deposits` — list of initial deposits
- `params` — governance parameters
- `constitution` — optional constitution text

## Events

The module emits events defined by the upstream `cosmos-sdk/x/gov` module:

| Event Type | Attributes |
|------------|------------|
| `submit_proposal` | `proposal_id`, `proposal_type`, `voting_period_start` |
| `proposal_deposit` | `proposal_id`, `depositor`, `amount` |
| `proposal_vote` | `proposal_id`, `voter`, `option` |
| `inactive_proposal` | `proposal_id`, `result` |
| `active_proposal` | `proposal_id`, `result` |
| `tally` | `proposal_id`, `yes`, `abstain`, `no`, `no_with_veto` |

## Client

Available through the standard Cosmos SDK gov CLI commands:

```bash
# Query proposals
safrochaind query gov proposal 1
safrochaind query gov proposals

# Governance actions
safrochaind tx gov submit-proposal --title="..." --description="..."
safrochaind tx gov deposit 1 100usafro
safrochaind tx gov vote 1 yes

# Query constitution
safrochaind query gov constitution
```
