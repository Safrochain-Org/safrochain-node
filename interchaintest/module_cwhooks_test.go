package interchaintest

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	helpers "github.com/Safrochain_Org/safrochain/tests/interchaintest/helpers"
)

// TestSafrochainCwHooks
// from x/cw-hooks/keeper/msg_server_test.go -> TestContractExecution
func TestSafrochainCwHooks(t *testing.T) {
	t.Parallel()

	cfg := safrochainConfig

	// Base setup
	chains := CreateChainWithCustomConfig(t, 1, 0, cfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	safrochain := chains[0].(*cosmos.CosmosChain)

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000_000), safrochain, safrochain)
	user := users[0]

	// Upload & init contract payment to another address
	_, contractAddr := helpers.SetupContract(t, ctx, safrochain, user.KeyName(), "contracts/safrochain_staking_hooks_example.wasm", `{}`)

	// register staking contract (to be tested)
	helpers.RegisterCwHooksStaking(t, ctx, safrochain, user, contractAddr)
	sc := helpers.GetCwHooksStakingContracts(t, ctx, safrochain)
	require.Equal(t, 1, len(sc))
	require.Equal(t, contractAddr, sc[0])

	// Validate that governance contract is added
	helpers.RegisterCwHooksGovernance(t, ctx, safrochain, user, contractAddr)
	gc := helpers.GetCwHooksGovernanceContracts(t, ctx, safrochain)
	require.Equal(t, 1, len(gc))
	require.Equal(t, contractAddr, gc[0])

	// Perform a Staking Action
	vals := helpers.GetValidators(t, ctx, safrochain)
	valoper := vals.Validators[0].OperatorAddress

	stakeAmt := 1_000_000
	helpers.StakeTokens(t, ctx, safrochain, user, valoper, fmt.Sprintf("%d%s", stakeAmt, safrochain.Config().Denom))

	// Query the smart contract to validate it saw the fire-and-forget update
	res := helpers.GetCwStakingHookLastDelegationChange(t, ctx, safrochain, contractAddr, user.FormattedAddress())
	require.Equal(t, valoper, res.Data.ValidatorAddress)
	require.Equal(t, user.FormattedAddress(), res.Data.DelegatorAddress)
	require.Equal(t, fmt.Sprintf("%d.000000000000000000", stakeAmt), res.Data.Shares)

	// HIGH GAS TEST
	// Setup a high gas contract
	_, contractAddr = helpers.SetupContract(t, ctx, safrochain, user.KeyName(), "contracts/safrochain_staking_hooks_high_gas_example.wasm", `{}`)

	// Register staking contract
	helpers.RegisterCwHooksStaking(t, ctx, safrochain, user, contractAddr)
	sc = helpers.GetCwHooksStakingContracts(t, ctx, safrochain)
	require.Equal(t, 2, len(sc))

	// Perform a Staking Action
	stakeAmt = 1_000_000
	helpers.StakeTokens(t, ctx, safrochain, user, valoper, fmt.Sprintf("%d%s", stakeAmt, safrochain.Config().Denom))

	// Query the smart contract, should panic and not update value
	res = helpers.GetCwStakingHookLastDelegationChange(t, ctx, safrochain, contractAddr, user.FormattedAddress())
	require.Equal(t, "", res.Data.ValidatorAddress)
	require.Equal(t, "", res.Data.DelegatorAddress)
	require.Equal(t, "", res.Data.Shares)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
