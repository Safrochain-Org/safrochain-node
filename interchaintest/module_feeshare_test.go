package interchaintest

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	helpers "github.com/Safrochain_Org/safrochain/tests/interchaintest/helpers"
)

// TestSafrochainFeeShare ensures the feeshare module register and execute sharing functions work properly on smart contracts.
func TestSafrochainFeeShare(t *testing.T) {
	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	safrochain := chains[0].(*cosmos.CosmosChain)

	nativeDenom := safrochain.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000), safrochain, safrochain)
	granter := users[0]
	grantee := users[1]
	feeRcvAddr := "safrochain1v75wlkccpv7le3560zw32v2zjes5n0e7csr4qh"

	// Upload & init contract payment to another address
	_, contractAddr := helpers.SetupContract(t, ctx, safrochain, granter.KeyName(), "contracts/cw_template.wasm", `{"count":0}`)

	// register contract to a random address (since we are the creator, though not the admin)
	helpers.RegisterFeeShare(t, ctx, safrochain, granter, contractAddr, feeRcvAddr)
	if balance, err := safrochain.GetBalance(ctx, feeRcvAddr, nativeDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 0 {
		t.Fatal("balance not 0")
	}

	// execute with a 10000 fee (so 5000 denom should be in the contract now with 50% feeshare default)
	_, err := helpers.ExecuteMsgWithFeeReturn(t, ctx, safrochain, granter, contractAddr, "", "10000"+nativeDenom, `{"increment":{}}`)
	if err != nil {
		t.Fatal(err)
	}

	// check balance of nativeDenom now
	if balance, err := safrochain.GetBalance(ctx, feeRcvAddr, nativeDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 5000 {
		t.Fatal("balance not 5,000. it is ", balance, nativeDenom)
	}

	// Test authz message execution:
	// Grant contract execute permission to grantee
	helpers.ExecuteAuthzGrantMsg(t, ctx, safrochain, granter, grantee, "/cosmos.authz.v1beta1.MsgExec")

	// Execute authz msg as grantee
	helpers.ExecuteAuthzExecMsgWithFee(t, ctx, safrochain, grantee, contractAddr, "", "10000"+nativeDenom, `{"increment":{}}`)

	// check balance of nativeDenom now
	if balance, err := safrochain.GetBalance(ctx, feeRcvAddr, nativeDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 10000 {
		t.Fatal("balance not 10,000. it is ", balance, nativeDenom)
	}

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
