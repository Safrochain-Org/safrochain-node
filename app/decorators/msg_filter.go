package decorators

import (
	"fmt"

	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// MsgFilterDecorator defines an AnteHandler decorator for the v9 upgrade that
// provide height-gated message filtering acceptance.
type MsgFilterDecorator struct{}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message
// that is blocked.
// Right now, we block MsgTimeoutOnClose due to incorrect behavior that could occur if a packet is re-enabled.
func (MsgFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	invalid, err := hasInvalidMsgs(tx.GetMsgs(), 0)
	if err != nil {
		return ctx, err
	}
	if invalid {
		currHeight := ctx.BlockHeight()
		return ctx, fmt.Errorf("tx contains unsupported message types at height %d", currHeight)
	}

	return next(ctx, tx, simulate)
}

// hasInvalidMsgs walks msgs and reports whether any are on the deny-list.
//
// SAF-12: also recurse into authz.MsgExec messages so the deny-list cannot
// be trivially bypassed by wrapping a blocked message in an authz exec.
// Recursion is bounded by MaxAuthzNestedMsgsDepth (SAF-05).
func hasInvalidMsgs(msgs []sdk.Msg, depth uint8) (bool, error) {
	if err := CheckAuthzDepth(depth); err != nil {
		return false, err
	}

	for _, msg := range msgs {
		if _, ok := msg.(*ibcchanneltypes.MsgTimeoutOnClose); ok {
			return true, nil
		}

		if execMsg, ok := msg.(*authz.MsgExec); ok {
			innerMsgs, err := execMsg.GetMessages()
			if err != nil {
				return false, err
			}
			invalid, err := hasInvalidMsgs(innerMsgs, depth+1)
			if err != nil {
				return false, err
			}
			if invalid {
				return true, nil
			}
		}
	}

	return false, nil
}
