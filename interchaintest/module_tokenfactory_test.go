package interchaintest

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	helpers "github.com/Safrochain_Org/safrochain/tests/interchaintest/helpers"
)

// TestSafrochainTokenFactory ensures the tokenfactory module & bindings work properly.
func TestSafrochainTokenFactory(t *testing.T) {
	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	safrochain := chains[0].(*cosmos.CosmosChain)
	t.Log("safrochain.GetHostRPCAddress()", safrochain.GetHostRPCAddress())

	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000), safrochain, safrochain)
	user := users[0]
	uaddr := user.FormattedAddress()

	user2 := users[1]
	uaddr2 := user2.FormattedAddress()

	tfDenom := helpers.CreateTokenFactoryDenom(t, ctx, safrochain, user, "ictestdenom", fmt.Sprintf("0%s", Denom))
	t.Log("tfDenom", tfDenom)

	// mint
	helpers.MintTokenFactoryDenom(t, ctx, safrochain, user, 100, tfDenom)
	t.Log("minted tfDenom to user")
	if balance, err := safrochain.GetBalance(ctx, uaddr, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 100 {
		t.Fatal("balance not 100")
	}

	// mint-to
	helpers.MintToTokenFactoryDenom(t, ctx, safrochain, user, user2, 70, tfDenom)
	t.Log("minted tfDenom to user")
	if balance, err := safrochain.GetBalance(ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 70 {
		t.Fatal("balance not 70")
	}

	// This allows the uaddr here to mint tokens on behalf of the contract. Typically you only allow a contract here, but this is testing.
	coreInitMsg := fmt.Sprintf(`{"allowed_mint_addresses":["%s"],"existing_denoms":["%s"]}`, uaddr, tfDenom)
	_, coreTFContract := helpers.SetupContract(t, ctx, safrochain, user.KeyName(), "contracts/tokenfactory_core.wasm", coreInitMsg)
	t.Log("coreContract", coreTFContract)

	// change admin to the contract
	helpers.TransferTokenFactoryAdmin(t, ctx, safrochain, user, coreTFContract, tfDenom)

	// ensure the admin is the contract
	admin := helpers.GetTokenFactoryAdmin(t, ctx, safrochain, tfDenom)
	t.Log("admin", admin)
	if admin != coreTFContract {
		t.Fatal("admin not coreTFContract. Did not properly transfer.")
	}

	// Mint on the contract for the user to ensure mint bindings work.
	mintMsg := fmt.Sprintf(`{"mint":{"address":"%s","denom":[{"denom":"%s","amount":"31"}]}}`, uaddr2, tfDenom)
	if _, err := safrochain.ExecuteContract(ctx, user.KeyName(), coreTFContract, mintMsg); err != nil {
		t.Fatal(err)
	}

	// ensure uaddr2 has 31+70 = 101
	if balance, err := safrochain.GetBalance(ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 101 {
		t.Fatal("balance not 101")
	}

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
