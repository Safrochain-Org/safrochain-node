package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

func TestMsgCreateDenomValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		desc  string
		msg   types.MsgCreateDenom
		valid bool
	}{
		{"valid", types.MsgCreateDenom{Sender: tfTestCreator, Subdenom: "bitcoin"}, true},
		{"empty sender", types.MsgCreateDenom{Subdenom: "bitcoin"}, false},
		{"invalid sender", types.MsgCreateDenom{Sender: "notanaddress", Subdenom: "bitcoin"}, false},
		{"empty subdenom", types.MsgCreateDenom{Sender: tfTestCreator}, false},
		{"subdenom too long", types.MsgCreateDenom{Sender: tfTestCreator, Subdenom: strings.Repeat("a", types.MaxSubdenomLength+1)}, false},
		{"invalid subdenom chars", types.MsgCreateDenom{Sender: tfTestCreator, Subdenom: "bit coin"}, false},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
