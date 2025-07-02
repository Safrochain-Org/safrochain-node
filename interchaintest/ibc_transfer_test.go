package interchaintest

import (
	"context"
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

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

// TestSafrochainGaiaIBCTransfer spins up a Safrochain and Gaia network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from Safrochain->Gaia and then back from Gaia->Safrochain.
func TestSafrochainGaiaIBCTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with Safrochain and Gaia
	numVals := 1
	numFullNodes := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "safrochain",
			ChainConfig:   safrochainConfig,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "ibc",
			ChainConfig:   ibcConfig,
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

	safrochain, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	relayerType, relayerName := ibc.CosmosRly, "rly"
	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		relayerType,
		zaptest.NewLogger(t),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	r := rf.Build(t, client, network)

	ic := interchaintest.NewInterchain().
		AddChain(safrochain).
		AddChain(gaia).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  safrochain,
			Chain2:  gaia,
			Relayer: r,
			Path:    path,
		})

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Create some user accounts on both chains
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, safrochain, gaia)

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, safrochain, gaia)
	require.NoError(t, err)

	// Get our Bech32 encoded user addresses
	safrochainUser, gaiaUser := users[0], users[1]

	safrochainUserAddr := safrochainUser.FormattedAddress()
	gaiaUserAddr := gaiaUser.FormattedAddress()

	// Get original account balances
	safrochainOrigBal, err := safrochain.GetBalance(ctx, safrochainUserAddr, safrochain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletAmount, safrochainOrigBal)

	gaiaOrigBal, err := gaia.GetBalance(ctx, gaiaUserAddr, gaia.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletAmount, gaiaOrigBal)

	// Compose an IBC transfer and send from Safrochain -> Gaia
	transferAmount := math.NewInt(1_000)
	transfer := ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   safrochain.Config().Denom,
		Amount:  transferAmount,
	}

	channel, err := ibc.GetTransferChannel(ctx, r, eRep, safrochain.Config().ChainID, gaia.Config().ChainID)
	require.NoError(t, err)

	safrochainHeight, err := safrochain.Height(ctx)
	require.NoError(t, err)

	transferTx, err := safrochain.SendIBCTransfer(ctx, channel.ChannelID, safrochainUserAddr, transfer, ibc.TransferOptions{})
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

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, safrochain, safrochainHeight, safrochainHeight+50, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 10, safrochain)
	require.NoError(t, err)

	// Get the IBC denom for usaf on Gaia
	safrochainTokenDenom := transfertypes.GetPrefixedDenom(channel.Counterparty.PortID, channel.Counterparty.ChannelID, safrochain.Config().Denom)
	safrochainIBCDenom := transfertypes.ParseDenomTrace(safrochainTokenDenom).IBCDenom()

	// Assert that the funds are no longer present in user acc on Safrochain and are in the user acc on Gaia
	safrochainUpdateBal, err := safrochain.GetBalance(ctx, safrochainUserAddr, safrochain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, safrochainOrigBal.Sub(transferAmount), safrochainUpdateBal)

	gaiaUpdateBal, err := gaia.GetBalance(ctx, gaiaUserAddr, safrochainIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, gaiaUpdateBal)

	// Compose an IBC transfer and send from Gaia -> Safrochain
	transfer = ibc.WalletAmount{
		Address: safrochainUserAddr,
		Denom:   safrochainIBCDenom,
		Amount:  transferAmount,
	}

	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	transferTx, err = gaia.SendIBCTransfer(ctx, channel.Counterparty.ChannelID, gaiaUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on Safrochain and not on Gaia
	safrochainUpdateBal, err = safrochain.GetBalance(ctx, safrochainUserAddr, safrochain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, safrochainOrigBal, safrochainUpdateBal)

	gaiaUpdateBal, err = gaia.GetBalance(ctx, gaiaUserAddr, safrochainIBCDenom)
	require.NoError(t, err)
	require.Equal(t, int64(0), gaiaUpdateBal.Int64())
}
