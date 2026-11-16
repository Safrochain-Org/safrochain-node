package keepers_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/app/keepers"
)

// The pointer type is what is wired as ICS4Middleware.Hooks.
var (
	_ ibchooks.SendPacketOverrideHooks              = (*keepers.SafeWasmHooks)(nil)
	_ ibchooks.OnRecvPacketOverrideHooks            = (*keepers.SafeWasmHooks)(nil)
	_ ibchooks.OnAcknowledgementPacketOverrideHooks = (*keepers.SafeWasmHooks)(nil)
	_ ibchooks.OnTimeoutPacketOverrideHooks         = (*keepers.SafeWasmHooks)(nil)
)

func TestValidateIBCCallback(t *testing.T) {
	sender := sdk.AccAddress([]byte("sender______________")).String()
	other := sdk.AccAddress([]byte("other_______________")).String()

	packet := func(sender, memo string) []byte {
		bz, err := json.Marshal(transfertypes.FungibleTokenPacketData{
			Denom: "usaf", Amount: "1", Sender: sender, Receiver: "r", Memo: memo,
		})
		require.NoError(t, err)
		return bz
	}

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"not ics20", []byte("garbage"), false},
		{"no memo", packet(sender, ""), false},
		{"non-json memo", packet(sender, "hello"), false},
		{"memo without callback", packet(sender, `{"wasm":{}}`), false},
		{"callback equals sender", packet(sender, `{"ibc_callback":"`+sender+`"}`), false},
		{"callback is another contract", packet(sender, `{"ibc_callback":"`+other+`"}`), true},
		{"callback not a string", packet(sender, `{"ibc_callback":123}`), true},
		{"callback null", packet(sender, `{"ibc_callback":null}`), true},
		{"callback invalid bech32", packet(sender, `{"ibc_callback":"nope"}`), true},
		{"sender not bech32", packet("nope", `{"ibc_callback":"`+sender+`"}`), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := keepers.ValidateIBCCallback(tc.data)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
