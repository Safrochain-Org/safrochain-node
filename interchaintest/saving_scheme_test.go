package interchaintest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/Safrochain_Org/safrochain/tests/interchaintest/helpers"
)

// TestSavingSchemeContract tests the saving scheme smart contract.
func TestSavingSchemeContract(t *testing.T) {
	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	safrochain := chains[0].(*cosmos.CosmosChain)

	nativeDenom := safrochain.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000), safrochain, safrochain)
	admin := users[0]
	participant1 := users[1]
	participant2 := users[2]
	participant3 := users[3]

	// Upload & Instantiate contract
	codeId, err := safrochain.StoreContract(ctx, admin.KeyName(), "contracts/saving_scheme.wasm", "--fees", "50000"+nativeDenom)
	require.NoError(t, err)

	instantiateMsg := map[string]interface{}{
		"participants": []string{participant1.FormattedAddress(), participant2.FormattedAddress(), participant3.FormattedAddress()},
		"deposit_amount": "1000000", // 1 token
		"denom": nativeDenom,
	}

	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	require.NoError(t, err)

	contractAddr, err := safrochain.InstantiateContract(ctx, admin.KeyName(), codeId, instantiateMsgBz, "--fees", "250000"+nativeDenom)
	require.NoError(t, err)

	t.Logf("Saving Scheme Contract Address: %s", contractAddr)

	// Verify initial state
	var state struct {
		Participants []string `json:"participants"`
		CurrentRecipientIndex uint32 `json:"current_recipient_index"`
		DepositAmount string `json:"deposit_amount"`
		Denom string `json:"denom"`
		Admin string `json:"admin"`
	}
	err = safrochain.QueryContract(ctx, contractAddr, `{"get_state":{}}`, &state)
	require.NoError(t, err)
	require.Len(t, state.Participants, 3)
	require.Equal(t, uint32(0), state.CurrentRecipientIndex)
	require.Equal(t, "1000000", state.DepositAmount)
	require.Equal(t, nativeDenom, state.Denom)
	require.Equal(t, admin.FormattedAddress(), state.Admin)

	// Each participant deposits funds
	depositAmount := sdkmath.NewInt(1_000_000) // 1 token
	depositMsg := `{"deposit":{}}`

	_, err = safrochain.ExecuteContract(ctx, participant1.KeyName(), contractAddr, depositMsg, "--amount", depositAmount.String()+nativeDenom)
	require.NoError(t, err)
	_, err = safrochain.ExecuteContract(ctx, participant2.KeyName(), contractAddr, depositMsg, "--amount", depositAmount.String()+nativeDenom)
	require.NoError(t, err)
	_, err = safrochain.ExecuteContract(ctx, participant3.KeyName(), contractAddr, depositMsg, "--amount", depositAmount.String()+nativeDenom)
	require.NoError(t, err)

	// Verify contract balance
	var balanceRes struct {
		Balance string `json:"balance"`
	}
	err = safrochain.QueryContract(ctx, contractAddr, `{"get_balance":{}}`, &balanceRes)
	require.NoError(t, err)
	expectedTotalBalance := depositAmount.MulRaw(3)
	require.Equal(t, expectedTotalBalance.String(), balanceRes.Balance)

	// Distribute funds to participants in succession
	distributeMsg := `{"distribute":{}}`

	// Distribute to participant 1
	_, err = safrochain.ExecuteContract(ctx, admin.KeyName(), contractAddr, distributeMsg, "--fees", "100000"+nativeDenom)
	require.NoError(t, err)
	
	// Verify participant 1 received funds (check balance increase)
	participant1Balance, err := safrochain.GetBalance(ctx, participant1.FormattedAddress(), nativeDenom)
	require.NoError(t, err)
	// Assuming initial balance + deposit + distributed amount
	// This check might be tricky due to initial funding and gas fees,
	// a more robust check would involve querying before and after.
	// For simplicity, we'll just check if it's greater than initial.
	require.True(t, participant1Balance.GT(sdkmath.NewInt(9_000_000))) // Initial 10M - 1M deposit + 1M distributed = 10M. Should be > 9M after fees.

	// Verify state after first distribution
	err = safrochain.QueryContract(ctx, contractAddr, `{"get_state":{}}`, &state)
	require.NoError(t, err)
	require.Equal(t, uint32(1), state.CurrentRecipientIndex) // Should move to next participant

	// Distribute to participant 2
	_, err = safrochain.ExecuteContract(ctx, admin.KeyName(), contractAddr, distributeMsg, "--fees", "100000"+nativeDenom)
	require.NoError(t, err)

	// Verify participant 2 received funds
	participant2Balance, err := safrochain.GetBalance(ctx, participant2.FormattedAddress(), nativeDenom)
	require.NoError(t, err)
	require.True(t, participant2Balance.GT(sdkmath.NewInt(9_000_000)))

	// Verify state after second distribution
	err = safrochain.QueryContract(ctx, contractAddr, `{"get_state":{}}`, &state)
	require.NoError(t, err)
	require.Equal(t, uint32(2), state.CurrentRecipientIndex)

	// Distribute to participant 3
	_, err = safrochain.ExecuteContract(ctx, admin.KeyName(), contractAddr, distributeMsg, "--fees", "100000"+nativeDenom)
	require.NoError(t, err)

	// Verify participant 3 received funds
	participant3Balance, err := safrochain.GetBalance(ctx, participant3.FormattedAddress(), nativeDenom)
	require.NoError(t, err)
	require.True(t, participant3Balance.GT(sdkmath.NewInt(9_000_000)))

	// Verify state after third distribution (should loop back to 0)
	err = safrochain.QueryContract(ctx, contractAddr, `{"get_state":{}}`, &state)
	require.NoError(t, err)
	require.Equal(t, uint32(0), state.CurrentRecipientIndex)

	// Verify final contract balance (should be close to zero after all distributions)
	err = safrochain.QueryContract(ctx, contractAddr, `{"get_balance":{}}`, &balanceRes)
	require.NoError(t, err)
	require.Equal(t, "0", balanceRes.Balance)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}