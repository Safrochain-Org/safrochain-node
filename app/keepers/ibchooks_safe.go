package keepers

import (
	"bytes"
	"encoding/json"
	"fmt"

	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile-time assertions that every hook implemented by the embedded
// ibchooks.WasmHooks (and the one overridden here) is exposed by SafeWasmHooks,
// so the ICS4/IBC middleware type assertions keep matching.
var (
	_ ibchooks.SendPacketOverrideHooks              = SafeWasmHooks{}
	_ ibchooks.OnRecvPacketOverrideHooks            = SafeWasmHooks{}
	_ ibchooks.OnAcknowledgementPacketOverrideHooks = SafeWasmHooks{}
	_ ibchooks.OnTimeoutPacketOverrideHooks         = SafeWasmHooks{}
)

// SafeWasmHooks wraps ibchooks.WasmHooks and only allows the packet sender to
// register an `ibc_callback` contract in an ICS-20 memo. The upstream hook
// stores the callback for any contract named in the memo, which lets anyone
// (including through packet-forward-middleware `next` memos) trigger
// `ibc_lifecycle_complete` sudo calls on arbitrary contracts.
type SafeWasmHooks struct {
	ibchooks.WasmHooks
}

// NewSafeWasmHooks wraps the given WasmHooks.
func NewSafeWasmHooks(h ibchooks.WasmHooks) SafeWasmHooks {
	return SafeWasmHooks{WasmHooks: h}
}

// SendPacketOverride validates any `ibc_callback` in the ICS-20 memo before
// delegating to the embedded implementation, which then returns the real packet
// sequence for valid callbacks.
func (h SafeWasmHooks) SendPacketOverride(
	i ibchooks.ICS4Middleware,
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight ibcclienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	if err := ValidateIBCCallback(data); err != nil {
		return 0, err
	}
	return h.WasmHooks.SendPacketOverride(i, ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// ValidateIBCCallback checks the packet data of an outgoing packet. Non ICS-20
// packets and ICS-20 packets without an `ibc_callback` memo key are accepted.
// If the key is present it must be a string holding a valid bech32 address that
// equals the packet sender.
func ValidateIBCCallback(data []byte) error {
	var ics20data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(data, &ics20data); err != nil {
		return nil // not an ICS-20 packet
	}
	if len(ics20data.Memo) == 0 {
		return nil
	}

	var metadata map[string]json.RawMessage
	if err := json.Unmarshal([]byte(ics20data.Memo), &metadata); err != nil {
		return nil // memo is not a JSON object, hooks ignore it
	}
	raw, found := metadata[ibchookstypes.IBCCallbackKey]
	if !found {
		return nil
	}

	var contract string
	if err := json.Unmarshal(raw, &contract); err != nil {
		return fmt.Errorf("%s must be a string", ibchookstypes.IBCCallbackKey)
	}
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return fmt.Errorf("%s is not a valid bech32 address: %w", ibchookstypes.IBCCallbackKey, err)
	}
	senderAddr, err := sdk.AccAddressFromBech32(ics20data.Sender)
	if err != nil {
		return fmt.Errorf("packet sender is not a valid bech32 address: %w", err)
	}
	if !bytes.Equal(contractAddr, senderAddr) {
		return fmt.Errorf("%s %s must be the packet sender %s", ibchookstypes.IBCCallbackKey, contract, ics20data.Sender)
	}
	return nil
}
