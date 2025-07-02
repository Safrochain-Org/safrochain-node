package setup

import (
	"encoding/json"
	"os"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	abci "github.com/cometbft/cometbft/abci/types"

	db "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	safrochainapp "github.com/Safrochain_Org/safrochain/app"
	"github.com/Safrochain_Org/safrochain/testutil/common"
)

// SetupTestingAppWithLevelDb initializes a new App intended for testing,
// with LevelDB as a db.
func SetupTestingAppWithLevelDB(isCheckTx bool) (app *safrochainapp.App, cleanupFn func()) { // nolint:revive
	dir, err := os.MkdirTemp(os.TempDir(), "safrochain_leveldb_testing")
	if err != nil {
		panic(err)
	}
	leveldb, err := db.NewGoLevelDB("safrochain_leveldb_testing", dir, nil)
	if err != nil {
		panic(err)
	}

	app = safrochainapp.New(log.NewNopLogger(), leveldb, nil, true, dir, sims.EmptyAppOptions{}, []wasmkeeper.Option{}, baseapp.SetChainID("safrochain-1"))
	if !isCheckTx {
		valSet := common.GenerateValidatorSet(1)
		genesisState := safrochainapp.NewDefaultGenesisState(app.AppCodec())
		genAccs := common.GenerateGenesisAccounts(1)
		authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
		genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)
		balances := []banktypes.Balance{}
		genesisState = common.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		_, err = app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
				ChainId:         "safrochain-1",
			},
		)
		if err != nil {
			panic(err)
		}
	}

	cleanupFn = func() {
		err = leveldb.Close()
		if err != nil {
			panic(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}

	return app, cleanupFn
}
