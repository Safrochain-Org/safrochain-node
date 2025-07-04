package interchaintest

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	helpers "github.com/Safrochain_Org/safrochain/tests/interchaintest/helpers"
)

// TestSafrochainUnityContractDeploy test to ensure the contract withdraw function works as expected on chain.
// - https://github.com/Safrochain_Org/cw-unity-prop
func TestSafrochainUnityContractDeploy(t *testing.T) {
	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	safrochain := chains[0].(*cosmos.CosmosChain)
	nativeDenom := safrochain.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000), safrochain, safrochain)
	user := users[0]
	withdrawUser := users[1]
	withdrawAddr := withdrawUser.FormattedAddress()

	// TEST DEPLOY (./scripts/deploy_ci.sh)
	// Upload & init unity contract with no admin in test mode
	msg := fmt.Sprintf(`{"native_denom":"%s","withdraw_address":"%s","withdraw_delay_in_days":28}`, nativeDenom, withdrawAddr)
	_, contractAddr := helpers.SetupContract(t, ctx, safrochain, user.KeyName(), "contracts/cw_unity_prop.wasm", msg)
	t.Log("testing Unity contractAddr", contractAddr)

	// Execute to start the withdrawal countdown
	safrochain.ExecuteContract(ctx, withdrawUser.KeyName(), contractAddr, `{"start_withdraw":{}}`)

	// make a query with GetUnityContractWithdrawalReadyTime
	res := helpers.GetUnityContractWithdrawalReadyTime(t, ctx, safrochain, contractAddr)
	t.Log("WithdrawalReadyTimestamp", res.Data.WithdrawalReadyTimestamp)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
