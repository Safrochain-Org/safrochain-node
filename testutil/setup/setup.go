package setup

import (
	"encoding/json"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	abci "github.com/cometbft/cometbft/abci/types"

	cosmosdb "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	safrochainapp "github.com/Safrochain_Org/safrochain/app"
	"github.com/Safrochain_Org/safrochain/testutil/common"
)

var defaultGenesisStateBytes = []byte{}

// Setup initializes a new App.
func Setup(isCheckTx bool, homePath, chainID string, t ...*testing.T) *safrochainapp.App {
	db := cosmosdb.NewMemDB()
	var (
		l       log.Logger
		appOpts servertypes.AppOptions
	)
	if common.IsDebugLogEnabled() {
		appOpts = common.DebugAppOptions{}
	} else {
		appOpts = sims.EmptyAppOptions{}
	}

	if len(t) > 0 {
		testEnv := t[0]
		testEnv.Log("Using test environment logger")
		l = log.NewTestLogger(testEnv)
	} else {
		l = log.NewNopLogger()
	}

	app := safrochainapp.New(
		l,
		db,
		nil,
		true,
		homePath,
		appOpts,
		[]wasmkeeper.Option{},
		baseapp.SetChainID(chainID),
	)
	if !isCheckTx {
		if len(defaultGenesisStateBytes) == 0 {
			var err error
			valSet := common.GenerateValidatorSet(1)
			genesisState := safrochainapp.NewDefaultGenesisState(app.AppCodec())
			genAccs := common.GenerateGenesisAccounts(1)
			authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
			genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)
			balances := []banktypes.Balance{}
			genesisState = common.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
			defaultGenesisStateBytes, err = json.Marshal(genesisState)
			if err != nil {
				panic(err)
			}
		}

		_, err := app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   defaultGenesisStateBytes,
				ChainId:         chainID,
			},
		)
		if err != nil {
			panic(err)
		}
	}

	return app
}
