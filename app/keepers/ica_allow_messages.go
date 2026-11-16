package keepers

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ICAHostAllowMessages is the least-privilege list of message type URLs that
// interchain accounts may execute on this chain through the ICA host module.
// Use it instead of the "*" wildcard when configuring ICA host params.
var ICAHostAllowMessages = []string{
	// bank
	sdk.MsgTypeURL(&banktypes.MsgSend{}),
	// staking
	sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{}),
	// distribution
	sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
	sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
	// governance
	sdk.MsgTypeURL(&govv1.MsgVote{}),
	sdk.MsgTypeURL(&govv1.MsgVoteWeighted{}),
	sdk.MsgTypeURL(&govv1.MsgDeposit{}),
	sdk.MsgTypeURL(&govv1beta1.MsgVote{}),
	sdk.MsgTypeURL(&govv1beta1.MsgVoteWeighted{}),
	// ibc transfer
	sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
	// wasm
	sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{}),
}
