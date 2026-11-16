package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	icagenesistypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/genesis/types"

	"github.com/Safrochain_Org/safrochain/app/keepers"
)

func TestDefaultIcaGenesisState(t *testing.T) {
	gs := defaultIcaGenesisState()
	require.NoError(t, gs.Validate())

	require.True(t, gs.HostGenesisState.Params.HostEnabled)
	require.Equal(t, keepers.ICAHostAllowMessages, gs.HostGenesisState.Params.AllowMessages)
	require.NotContains(t, gs.HostGenesisState.Params.AllowMessages, "*")
	require.Contains(t, gs.HostGenesisState.Params.AllowMessages, "/cosmos.bank.v1beta1.MsgSend")
	require.Contains(t, gs.HostGenesisState.Params.AllowMessages, "/cosmwasm.wasm.v1.MsgExecuteContract")
	require.Len(t, gs.HostGenesisState.Params.AllowMessages, 14)

	// controller params are left at the ibc-go defaults, not zeroed
	require.Equal(t, icagenesistypes.DefaultControllerGenesis().Params, gs.ControllerGenesisState.Params)

	// mutating the result must not alter the shared allow-list
	gs.HostGenesisState.Params.AllowMessages[0] = "x"
	require.NotEqual(t, "x", keepers.ICAHostAllowMessages[0])
}
