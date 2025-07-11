package interchaintest

import (
	"context"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/docker/docker/client"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	clocktypes "github.com/Safrochain_Org/safrochain/x/clock/types"
	feepaytypes "github.com/Safrochain_Org/safrochain/x/feepay/types"
	feesharetypes "github.com/Safrochain_Org/safrochain/x/feeshare/types"
	globalfeetypes "github.com/Safrochain_Org/safrochain/x/globalfee/types"
	tokenfactorytypes "github.com/Safrochain_Org/safrochain/x/tokenfactory/types"
)

var (
	VotingPeriod     = "10s"
	MaxDepositPeriod = "10s"
	Denom            = "usaf"

	SafrochainMainRepo                = "ghcr.io/Safrochain_Org/safrochain"
	safrochainRepo, safrochainVersion = GetDockerImageInfo()

	SafrochainImage = ibc.DockerImage{
		Repository: safrochainRepo,
		Version:    safrochainVersion,
		UIDGID:     "1025:1025",
	}

	// SDK v47 Genesis
	defaultGenesisKV = []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: Denom,
		},
		{
			Key:   "app_state.feepay.params.enable_feepay",
			Value: false,
		},
	}

	safrochainConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "safrochain",
		ChainID:             "safrochain-2",
		Images:              []ibc.DockerImage{SafrochainImage},
		Bin:                 "safrochaind",
		Bech32Prefix:        "safrochain",
		Denom:               Denom,
		CoinType:            "118",
		GasPrices:           fmt.Sprintf("0%s", Denom),
		GasAdjustment:       2.0,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ConfigFileOverrides: nil,
		EncodingConfig:      safrochainEncoding(),
		ModifyGenesis:       cosmos.ModifyGenesis(defaultGenesisKV),
	}

	ibcConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "ibc-chain",
		ChainID:             "ibc-1",
		Images:              []ibc.DockerImage{SafrochainImage},
		Bin:                 "safrochaind",
		Bech32Prefix:        "safrochain",
		Denom:               "usaf",
		CoinType:            "118",
		GasPrices:           fmt.Sprintf("0%s", Denom),
		GasAdjustment:       2.0,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ConfigFileOverrides: nil,
		EncodingConfig:      safrochainEncoding(),
		ModifyGenesis:       cosmos.ModifyGenesis(defaultGenesisKV),
	}

	genesisWalletAmount = sdkmath.NewInt(10_000_000)
)

func init() {
	sdk.GetConfig().SetBech32PrefixForAccount("safrochain", "safrochain")
	sdk.GetConfig().SetBech32PrefixForValidator("safrochainvaloper", "safrochain")
	sdk.GetConfig().SetBech32PrefixForConsensusNode("safrochainvalcons", "safrochain")
	sdk.GetConfig().SetCoinType(118)
}

// safrochainEncoding registers the Safrochain specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func safrochainEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feepaytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	globalfeetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	clocktypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

// CreateChain generates a new chain with a custom image (useful for upgrades)
func CreateChain(t *testing.T, numVals, numFull int, img ibc.DockerImage) []ibc.Chain {
	cfg := safrochainConfig
	cfg.Images = []ibc.DockerImage{img}
	return CreateChainWithCustomConfig(t, numVals, numFull, cfg)
}

// CreateThisBranchChain generates this branch's chain (ex: from the commit)
func CreateThisBranchChain(t *testing.T, numVals, numFull int) []ibc.Chain {
	return CreateChain(t, numVals, numFull, SafrochainImage)
}

func CreateChainWithCustomConfig(t *testing.T, numVals, numFull int, config ibc.ChainConfig) []ibc.Chain {
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "safrochain",
			ChainName:     "safrochain",
			Version:       config.Images[0].Version,
			ChainConfig:   config,
			NumValidators: &numVals,
			NumFullNodes:  &numFull,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// chain := chains[0].(*cosmos.CosmosChain)
	return chains
}

func BuildInitialChain(t *testing.T, chains []ibc.Chain) (*interchaintest.Interchain, context.Context, *client.Client, string) {
	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()

	for _, chain := range chains {
		ic = ic.AddChain(chain)
	}

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err := ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	return ic, ctx, client, network
}
