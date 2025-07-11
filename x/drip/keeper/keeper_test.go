package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/Safrochain_Org/safrochain/testutil"
	"github.com/Safrochain_Org/safrochain/x/drip/keeper"
	"github.com/Safrochain_Org/safrochain/x/drip/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	genesis types.GenesisState

	bankKeeper bankkeeper.Keeper

	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.genesis = *types.DefaultGenesisState()

	s.bankKeeper = s.App.AppKeepers.BankKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(s.App.AppKeepers.DripKeeper)
}
