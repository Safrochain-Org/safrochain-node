package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultBypassFeeMessages(t *testing.T) {
	require.Equal(t, []string{
		"/ibc.core.channel.v1.MsgRecvPacket",
		"/ibc.core.channel.v1.MsgAcknowledgement",
		"/ibc.core.channel.v1.MsgTimeout",
		"/ibc.core.client.v1.MsgUpdateClient",
	}, GetDefaultBypassFeeMessages())
}
