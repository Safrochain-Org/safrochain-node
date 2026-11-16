package v030_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Safrochain_Org/safrochain/app/keepers"
	v030 "github.com/Safrochain_Org/safrochain/app/upgrades/v030"
	"github.com/Safrochain_Org/safrochain/testutil"
	clocktypes "github.com/Safrochain_Org/safrochain/x/clock/types"
	cwhookstypes "github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
	driptypes "github.com/Safrochain_Org/safrochain/x/drip/types"
	globalfeetypes "github.com/Safrochain_Org/safrochain/x/globalfee/types"
)

// liveTestnetVersions is the x/upgrade module version map of safro-testnet-1
// (binary v0.1.0), queried read-only before the upgrade. Running the handler
// from this map exercises the exact migration path of the live network.
var liveTestnetVersions = module.VersionMap{
	"auth": 5, "authz": 2, "bank": 4, "capability": 1, "clock": 1, "consensus": 1,
	"crisis": 2, "cw-hooks": 1, "distribution": 3, "drip": 1, "evidence": 1,
	"feegrant": 2, "feeibc": 2, "feepay": 1, "feeshare": 2, "genutil": 1,
	"globalfee": 2, "gov": 5, "ibc": 6, "ibchooks": 1, "interchainaccounts": 3,
	"interchainquery": 2, "mint": 2, "nft": 1, "packetfowardmiddleware": 3,
	"params": 1, "slashing": 4, "staking": 5, "tokenfactory": 2, "transfer": 5,
	"upgrade": 2, "vesting": 1, "wasm": 4,
}

type UpgradeTestSuite struct {
	testutil.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// setLiveTestnetState puts the parameters that differ from the new defaults
// into the values observed on safro-testnet-1. Params are written directly to
// the stores because the new validation rejects some of them.
func (s *UpgradeTestSuite) setLiveTestnetState() {
	k := s.App.AppKeepers

	cp, err := k.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	cp.Block.MaxGas = -1
	s.Require().NoError(k.ConsensusParamsKeeper.ParamsStore.Set(s.Ctx, cp))

	cdc := s.App.AppCodec()

	clockStore := runtime.KVStoreAdapter(k.ClockKeeper.GetStoreService().OpenKVStore(s.Ctx))
	clockStore.Set(clocktypes.ParamsKey, cdc.MustMarshal(&clocktypes.Params{ContractGasLimit: 100_000}))

	dripStore := s.Ctx.KVStore(k.GetKey(driptypes.StoreKey))
	dripStore.Set(driptypes.ParamsKey, cdc.MustMarshal(&driptypes.Params{EnableDrip: true}))

	s.Require().NoError(k.GlobalFeeKeeper.SetParams(s.Ctx, globalfeetypes.Params{}))

	k.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.NewParams(true, []string{icahosttypes.AllowAllHostMsgs}))
}

// runUpgrade schedules the plan and executes it through the x/upgrade
// PreBlocker, exactly as the chain does at the upgrade height.
func (s *UpgradeTestSuite) runUpgrade(height int64) {
	k := s.App.AppKeepers
	s.Require().NoError(k.UpgradeKeeper.SetModuleVersionMap(s.Ctx, liveTestnetVersions))

	plan := upgradetypes.Plan{Name: v030.UpgradeName, Height: height}
	s.Require().NoError(k.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan))

	s.Ctx = s.Ctx.WithBlockHeight(height).WithHeaderInfo(s.Ctx.HeaderInfo())
	headerInfo := s.Ctx.HeaderInfo()
	headerInfo.Height = height
	s.Ctx = s.Ctx.WithHeaderInfo(headerInfo)

	s.Require().NotPanics(func() {
		_, err := s.App.PreBlocker(s.Ctx, &abci.RequestFinalizeBlock{Height: height})
		s.Require().NoError(err)
	})

	doneHeight, err := k.UpgradeKeeper.GetDoneHeight(s.Ctx, v030.UpgradeName)
	s.Require().NoError(err)
	s.Require().Equal(height, doneHeight)
}

func (s *UpgradeTestSuite) TestUpgradeFromLiveTestnetState() {
	s.Setup()
	s.setLiveTestnetState()

	s.runUpgrade(s.Ctx.BlockHeight() + 5)

	k := s.App.AppKeepers

	cp, err := k.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(v030.BlockMaxGas, cp.Block.MaxGas)

	s.Require().Equal(clocktypes.DefaultContractGasLimit, k.ClockKeeper.GetParams(s.Ctx).ContractGasLimit)
	s.Require().NoError(k.CWHooksKeeper.GetParams(s.Ctx).Validate())

	drip := k.DripKeeper.GetParams(s.Ctx)
	s.Require().False(drip.EnableDrip)
	s.Require().NoError(drip.Validate())

	bondDenom, err := k.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(
		sdk.NewDecCoins(sdk.NewDecCoinFromDec(bondDenom, math.LegacyMustNewDecFromStr(v030.MinGasPrice))),
		k.GlobalFeeKeeper.GetParams(s.Ctx).MinimumGasPrices,
	)

	ica := k.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().True(ica.HostEnabled)
	s.Require().Equal(keepers.ICAHostAllowMessages, ica.AllowMessages)

	// The upgraded state must pass genesis validation: export/import is the
	// disaster-recovery path.
	s.requireExportedGenesisIsValid()
}

// mainnetICAHostAllowMessages is the ICA host allow-list configured on
// safrochain-1 (queried read-only). It is not a wildcard, so the upgrade must
// keep it as is.
var mainnetICAHostAllowMessages = []string{
	"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgMultiSend",
	"/cosmos.staking.v1beta1.MsgDelegate", "/cosmos.staking.v1beta1.MsgUndelegate",
	"/cosmos.staking.v1beta1.MsgBeginRedelegate", "/cosmos.staking.v1beta1.MsgCancelUnbondingDelegation",
	"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward", "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission",
	"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress", "/cosmos.distribution.v1beta1.MsgFundCommunityPool",
	"/cosmos.gov.v1.MsgVote", "/cosmos.gov.v1.MsgVoteWeighted", "/cosmos.gov.v1.MsgDeposit", "/cosmos.gov.v1.MsgSubmitProposal",
	"/cosmos.gov.v1beta1.MsgVote", "/cosmos.gov.v1beta1.MsgVoteWeighted", "/cosmos.gov.v1beta1.MsgDeposit", "/cosmos.gov.v1beta1.MsgSubmitProposal",
	"/cosmos.authz.v1beta1.MsgGrant", "/cosmos.authz.v1beta1.MsgRevoke", "/cosmos.authz.v1beta1.MsgExec",
	"/cosmos.feegrant.v1beta1.MsgGrantAllowance", "/cosmos.feegrant.v1beta1.MsgRevokeAllowance",
	"/ibc.applications.transfer.v1.MsgTransfer",
	"/cosmwasm.wasm.v1.MsgExecuteContract", "/cosmwasm.wasm.v1.MsgInstantiateContract", "/cosmwasm.wasm.v1.MsgInstantiateContract2",
}

// setLiveMainnetState puts the parameters observed on safrochain-1 (binary
// v0.2.2) into the store: the module version map is identical to the
// testnet's, but every module parameter is already valid except the
// unlimited block gas.
func (s *UpgradeTestSuite) setLiveMainnetState() {
	k := s.App.AppKeepers

	cp, err := k.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	cp.Block.MaxGas = -1
	s.Require().NoError(k.ConsensusParamsKeeper.ParamsStore.Set(s.Ctx, cp))

	s.Require().NoError(k.ClockKeeper.SetParams(s.Ctx, clocktypes.Params{ContractGasLimit: 1_000_000}))
	s.Require().NoError(k.CWHooksKeeper.SetParams(s.Ctx, cwhookstypes.NewParams(1_000_000)))
	s.Require().NoError(k.DripKeeper.SetParams(s.Ctx, driptypes.Params{}))
	s.Require().NoError(k.GlobalFeeKeeper.SetParams(s.Ctx, globalfeetypes.Params{
		MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec("usaf", math.LegacyMustNewDecFromStr("0.05"))),
	}))
	k.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.NewParams(true, mainnetICAHostAllowMessages))
}

// From the mainnet state the upgrade only bounds the block gas limit; every
// other parameter is preserved.
func (s *UpgradeTestSuite) TestUpgradeFromLiveMainnetState() {
	s.Setup()
	s.setLiveMainnetState()
	k := s.App.AppKeepers

	before := struct {
		clock, cwhooks, drip, globalfee, tokenfactory, feeshare any
	}{
		k.ClockKeeper.GetParams(s.Ctx), k.CWHooksKeeper.GetParams(s.Ctx), k.DripKeeper.GetParams(s.Ctx),
		k.GlobalFeeKeeper.GetParams(s.Ctx), k.TokenFactoryKeeper.GetParams(s.Ctx), k.FeeShareKeeper.GetParams(s.Ctx),
	}

	s.runUpgrade(s.Ctx.BlockHeight() + 5)

	cp, err := k.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(v030.BlockMaxGas, cp.Block.MaxGas)

	s.Require().Equal(before.clock, any(k.ClockKeeper.GetParams(s.Ctx)))
	s.Require().Equal(before.cwhooks, any(k.CWHooksKeeper.GetParams(s.Ctx)))
	s.Require().Equal(before.drip, any(k.DripKeeper.GetParams(s.Ctx)))
	s.Require().Equal(before.globalfee, any(k.GlobalFeeKeeper.GetParams(s.Ctx)))
	s.Require().Equal(before.tokenfactory, any(k.TokenFactoryKeeper.GetParams(s.Ctx)))
	s.Require().Equal(before.feeshare, any(k.FeeShareKeeper.GetParams(s.Ctx)))
	s.Require().Equal(mainnetICAHostAllowMessages, k.ICAHostKeeper.GetParams(s.Ctx).AllowMessages)

	s.requireExportedGenesisIsValid()
}

// Values that are already valid are left untouched.
func (s *UpgradeTestSuite) TestUpgradeKeepsValidValues() {
	s.Setup()
	k := s.App.AppKeepers

	cp, err := k.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	cp.Block.MaxGas = 50_000_000
	s.Require().NoError(k.ConsensusParamsKeeper.ParamsStore.Set(s.Ctx, cp))

	custom := sdk.NewDecCoins(sdk.NewDecCoinFromDec("usaf", math.LegacyMustNewDecFromStr("0.1")))
	s.Require().NoError(k.GlobalFeeKeeper.SetParams(s.Ctx, globalfeetypes.Params{MinimumGasPrices: custom}))

	allow := []string{"/cosmos.bank.v1beta1.MsgSend"}
	k.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.NewParams(true, allow))

	s.runUpgrade(s.Ctx.BlockHeight() + 5)

	cp, err = k.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(int64(50_000_000), cp.Block.MaxGas)
	s.Require().Equal(custom, k.GlobalFeeKeeper.GetParams(s.Ctx).MinimumGasPrices)
	s.Require().Equal(allow, k.ICAHostKeeper.GetParams(s.Ctx).AllowMessages)
}

// requireExportedGenesisIsValid exports the genesis of every module the
// upgrade touches and validates it with the module's own ValidateGenesis:
// export/import is the disaster-recovery path, so the upgraded state must pass
// the new, stricter validation.
func (s *UpgradeTestSuite) requireExportedGenesisIsValid() {
	cdc := s.App.AppCodec()
	for _, name := range []string{"clock", "cw-hooks", "drip", "globalfee", "tokenfactory", "feeshare", "interchainaccounts"} {
		mod, ok := s.App.ModuleManager.Modules[name].(module.HasGenesis)
		s.Require().True(ok, "module %s has no genesis", name)

		bz := mod.ExportGenesis(s.Ctx, cdc)
		basic, ok := s.App.BasicModuleManager[name].(module.HasGenesisBasics)
		s.Require().True(ok, "module %s has no genesis basics", name)
		s.Require().NoError(basic.ValidateGenesis(cdc, s.App.TxConfig(), bz), "module %s", name)
	}
}

// Without the upgrade, the live parameters fail the new validation, so a
// genesis export of the live network could not be re-imported.
func (s *UpgradeTestSuite) TestLiveStateFailsNewValidationWithoutUpgrade() {
	s.Setup()
	s.setLiveTestnetState()

	cdc := s.App.AppCodec()
	for _, name := range []string{"clock", "drip"} {
		bz := s.App.ModuleManager.Modules[name].(module.HasGenesis).ExportGenesis(s.Ctx, cdc)
		err := s.App.BasicModuleManager[name].(module.HasGenesisBasics).ValidateGenesis(cdc, s.App.TxConfig(), bz)
		s.Require().Error(err, "module %s", name)
	}
}
