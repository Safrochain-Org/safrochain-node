package keepers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Safrochain_Org/safrochain/app/keepers"
)

func TestGetMaccPermsReturnsCopy(t *testing.T) {
	perms := keepers.GetMaccPerms()
	require.NotEmpty(t, perms)

	for name := range perms {
		perms[name] = append(perms[name], "mutated")
	}
	perms["injected"] = nil

	fresh := keepers.GetMaccPerms()
	require.NotContains(t, fresh, "injected")
	for _, v := range fresh {
		require.NotContains(t, v, "mutated")
	}
}
