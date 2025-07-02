package interchaintest

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	helpers "github.com/Safrochain_Org/safrochain/tests/interchaintest/helpers"
)

// TestSafrochainIBCHooks ensures the ibc-hooks middleware from osmosis works.
func TestSafrochainIBCHooks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with Safrochain and safrochain2
	numVals := 1
	numFullNodes := 0

	cfg2 := safrochainConfig.Clone()
	cfg2.Name = "safrochain-counterparty"
	cfg2.ChainID = "counterparty-2"

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "safrochain",
			ChainConfig:   safrochainConfig,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "safrochain",
			ChainConfig:   cfg2,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	const (
		path = "ibc-path"
	)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	client, network := interchaintest.DockerSetup(t)

	safrochain, safrochain2 := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	relayerType, relayerName := ibc.CosmosRly, "relay"

	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		relayerType,
		zaptest.NewLogger(t),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)

	r := rf.Build(t, client, network)

	ic := interchaintest.NewInterchain().
		AddChain(safrochain).
		AddChain(safrochain2).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  safrochain,
			Chain2:  safrochain2,
			Relayer: r,
			Path:    path,
		})

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Create some user accounts on both chains
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, safrochain, safrochain2)

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, safrochain, safrochain2)
	require.NoError(t, err)

	// Get our Bech32 encoded user addresses
	safrochainUser, safrochain2User := users[0], users[1]

	safrochainUserAddr := safrochainUser.FormattedAddress()
	// safrochain2UserAddr := safrochain2User.FormattedAddress()

	channel, err := ibc.GetTransferChannel(ctx, r, eRep, safrochain.Config().ChainID, safrochain2.Config().ChainID)
	require.NoError(t, err)

	err = r.StartRelayer(ctx, eRep, path)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	_, contractAddr := helpers.SetupContract(t, ctx, safrochain2, safrochain2User.KeyName(), "contracts/ibchooks_counter.wasm", `{"count":0}`)

	// do an ibc transfer through the memo to the other chain.
	transfer := ibc.WalletAmount{
		Address: contractAddr,
		Denom:   safrochain.Config().Denom,
		Amount:  math.NewInt(1),
	}

	memo := ibc.TransferOptions{
		Memo: fmt.Sprintf(`{"wasm":{"contract":"%s","msg":%s}}`, contractAddr, `{"increment":{}}`),
	}

	// Initial transfer. Account is created by the wasm execute is not so we must do this twice to properly set up
	transferTx, err := safrochain.SendIBCTransfer(ctx, channel.ChannelID, safrochainUser.KeyName(), transfer, memo)
	require.NoError(t, err)
	safrochainHeight, err := safrochain.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, safrochain, safrochainHeight-5, safrochainHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Second time, this will make the counter == 1 since the account is now created.
	transferTx, err = safrochain.SendIBCTransfer(ctx, channel.ChannelID, safrochainUser.KeyName(), transfer, memo)
	require.NoError(t, err)
	safrochainHeight, err = safrochain.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, safrochain, safrochainHeight-5, safrochainHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Get the address on the other chain's side
	addr := helpers.GetIBCHooksUserAddress(t, ctx, safrochain, channel.ChannelID, safrochainUserAddr)
	require.NotEmpty(t, addr)

	// Get funds on the receiving chain
	funds := helpers.GetIBCHookTotalFunds(t, ctx, safrochain2, contractAddr, addr)
	require.Equal(t, int(1), len(funds.Data.TotalFunds))

	var ibcDenom string
	for _, coin := range funds.Data.TotalFunds {
		if strings.HasPrefix(coin.Denom, "ibc/") {
			ibcDenom = coin.Denom
			break
		}
	}
	require.NotEmpty(t, ibcDenom)

	// ensure the count also increased to 1 as expected.
	count := helpers.GetIBCHookCount(t, ctx, safrochain2, contractAddr, addr)
	require.Equal(t, int64(1), count.Data.Count)
}
