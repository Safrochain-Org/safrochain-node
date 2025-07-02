package interchaintest

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	helpers "github.com/Safrochain_Org/safrochain/tests/interchaintest/helpers"
)

// TestSafrochainFeePay
func TestSafrochainFeePay(t *testing.T) {
	t.Parallel()

	cfg := safrochainConfig
	cfg.GasPrices = "0.0025usaf"

	// 0.002500000000000000
	coin := sdk.NewDecCoinFromDec(cfg.Denom, sdkmath.LegacyNewDecWithPrec(25, 4))
	cfg.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesisKV, []cosmos.GenesisKV{
		{
			Key:   "app_state.globalfee.params.minimum_gas_prices",
			Value: sdk.DecCoins{coin},
		},
		{
			// override default impl.
			Key:   "app_state.feepay.params.enable_feepay",
			Value: true,
		},
	}...))

	// Base setup
	chains := CreateChainWithCustomConfig(t, 1, 0, cfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	safrochain := chains[0].(*cosmos.CosmosChain)

	nativeDenom := safrochain.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000), safrochain, safrochain)
	admin := users[0]
	user := users[1]

	// Upload & init contract payment to another address
	codeId, err := safrochain.StoreContract(ctx, admin.KeyName(), "contracts/cw_template.wasm", "--fees", "50000usaf")
	if err != nil {
		t.Fatal(err)
	}

	contractAddr, err := safrochain.InstantiateContract(ctx, admin.KeyName(), codeId, `{"count":0}`, true)
	if err != nil {
		t.Fatal(err)
	}

	// Register contract for 0 fee usage (x amount of times)
	limit := 5
	balance := 1_000_000
	helpers.RegisterFeePay(t, ctx, safrochain, admin, contractAddr, limit)
	helpers.FundFeePayContract(t, ctx, safrochain, admin, contractAddr, strconv.Itoa(balance)+nativeDenom)

	beforeContract := helpers.GetFeePayContract(t, ctx, safrochain, contractAddr)
	t.Log("beforeContract", beforeContract)
	require.Equal(t, beforeContract.FeePayContract.Balance, strconv.Itoa(balance))
	require.Equal(t, beforeContract.FeePayContract.WalletLimit, strconv.Itoa(int(limit)))

	// execute against it from another account with enough fees (standard Tx)
	txHash, err := safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	beforeBal, err := safrochain.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// execute against it from another account and have the dev pay it
	txHash, err = safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	afterBal, err := safrochain.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// validate users balance did not change
	require.Equal(t, beforeBal, afterBal)

	// validate the contract balance went down
	afterContract := helpers.GetFeePayContract(t, ctx, safrochain, contractAddr)
	t.Log("afterContract", afterContract)
	require.Equal(t, afterContract.FeePayContract.Balance, strconv.Itoa(balance-500))

	uses := helpers.GetFeePayUses(t, ctx, safrochain, contractAddr, user.FormattedAddress())
	t.Log("uses", uses)
	require.Equal(t, uses.Uses, "1")

	// Instantiate a new contract
	contractAddr, err = safrochain.InstantiateContract(ctx, admin.KeyName(), codeId, `{"count":0}`, true)
	if err != nil {
		t.Fatal(err)
	}

	// Succeed - Test a regular CW contract with fees, regular sdk logic handles Tx
	txHash, err = safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	// Fail - Testing an unregistered contract with no fees, FeePay Tx logic will fail it due to not being registered
	txHash, err = safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.Error(t, err)
	fmt.Println("txHash", txHash)

	// Register the new contract with a limit of 1, fund contract
	helpers.RegisterFeePay(t, ctx, safrochain, admin, contractAddr, 1)
	helpers.FundFeePayContract(t, ctx, safrochain, admin, contractAddr, strconv.Itoa(balance)+nativeDenom)

	// Test the registered contract - with fees
	// Will succeed, routed through normal sdk because a fee was provided
	txHash, err = safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "500"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	// Before balance - should be the same as after balance (feepay covers fee)
	// Calculated before interacting with a registered contract to ensure the
	// contract covers the fee.
	beforeBal, err = safrochain.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// Test the registered FeePay contract - without providing fees
	txHash, err = safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	// After balance
	afterBal, err = safrochain.GetBalance(ctx, user.FormattedAddress(), nativeDenom)
	require.NoError(t, err)

	// Validate users balance did not change
	require.Equal(t, beforeBal, afterBal)

	// Test the fallback sdk route is triggered when the FeePay Tx fails
	// Fail - Test the registered contract - without fees, exceeded wallet limit
	txHash, err = safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--fees", "0"+nativeDenom)
	require.Error(t, err)
	fmt.Println("txHash", txHash)

	// Test the registered contract - without fees, but specified gas
	// Tx should succeed, because it uses the sdk fallback route
	txHash, err = safrochain.ExecuteContract(ctx, user.KeyName(), contractAddr, `{"increment":{}}`, "--gas", "200000")
	require.NoError(t, err)
	fmt.Println("txHash", txHash)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
