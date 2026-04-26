package decorators

import (
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MaxAuthzNestedMsgsDepth is the maximum supported nesting depth for
// authz.MsgExec wrappers when an ante decorator recursively walks a
// transaction's messages.
//
// SAF-05: deeply nested authz transactions previously caused unbounded
// recursion in the change-rate and fee-share decorators, which an attacker
// could exploit to crash a node via stack exhaustion. The limit is also
// applied by the IBC msg-filter (SAF-12) so blocked message types cannot
// be smuggled through arbitrarily deep authz wrappers.
const MaxAuthzNestedMsgsDepth uint8 = 5

// ErrAuthzNestedMsgsTooDeep is returned when a transaction contains more
// than MaxAuthzNestedMsgsDepth nested authz.MsgExec wrappers.
var ErrAuthzNestedMsgsTooDeep = errorsmod.Register(
	"safrochain/decorators",
	2,
	"authz nested messages exceed maximum supported depth",
)

// CheckAuthzDepth returns ErrAuthzNestedMsgsTooDeep when `depth` would
// exceed MaxAuthzNestedMsgsDepth. Decorators that recursively unwrap
// authz.MsgExec messages must call this helper before recursing.
func CheckAuthzDepth(depth uint8) error {
	if depth > MaxAuthzNestedMsgsDepth {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"authz nested messages exceed maximum supported depth of %d", MaxAuthzNestedMsgsDepth,
		)
	}
	return nil
}
