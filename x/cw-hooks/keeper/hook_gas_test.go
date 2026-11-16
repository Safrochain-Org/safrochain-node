package keeper_test

import (
	"encoding/json"

	_ "embed"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/keeper"
	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// highGasContract burns its whole gas budget in its staking hook handlers. See
// testdata/README.md.
//
//go:embed testdata/high_gas.wasm
var highGasContract []byte

// validatorCreatedMsg returns a valid after_validator_created hook message for
// the first genesis validator.
func (s *KeeperTestSuite) validatorCreatedMsg() []byte {
	validators, err := s.stakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(validators)

	bz, err := json.Marshal(keeper.SudoMsgAfterValidatorCreated{
		AfterValidatorCreated: keeper.NewValidator(validators[0]),
	})
	s.Require().NoError(err)

	return bz
}

func (s *KeeperTestSuite) registerStakingHookContract(wasm []byte) sdk.AccAddress {
	_, _, creator := testdata.KeyTestPubAddr()
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin(s.bondDenom(), sdkmath.NewInt(1_000_000))))

	contract := s.storeAndInstantiate(creator, wasm)
	s.registerHook(contract, true, false)

	return contract
}

// Gas consumed by hook contracts is charged to the operation that triggered
// the hook instead of running on a free, unaccounted meter.
func (s *KeeperTestSuite) TestHookGasIsChargedToCaller() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	s.registerStakingHookContract(highGasContract)
	msg := s.validatorCreatedMsg()

	gasLimit := s.App.AppKeepers.CWHooksKeeper.GetParams(s.Ctx).ContractGasLimit
	ctx := s.Ctx.WithGasMeter(storetypes.NewGasMeter(10 * gasLimit))

	s.App.AppKeepers.CWHooksKeeper.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msg)

	// The contract burns its whole budget, and that budget is charged to the
	// caller on top of the registry and params reads.
	consumed := ctx.GasMeter().GasConsumed()
	s.Require().GreaterOrEqual(consumed, gasLimit)
	s.Require().Less(consumed, gasLimit+20_000)
}

// A hook contract never gets more gas than the caller has left, so it cannot
// push the caller past its own gas limit.
func (s *KeeperTestSuite) TestHookGasIsCappedByCallerRemainingGas() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	contract := s.registerStakingHookContract(highGasContract)
	msg := s.validatorCreatedMsg()

	gasLimit := s.App.AppKeepers.CWHooksKeeper.GetParams(s.Ctx).ContractGasLimit
	callerLimit := gasLimit / 2
	ctx := s.Ctx.WithGasMeter(storetypes.NewGasMeter(callerLimit)).WithEventManager(sdk.NewEventManager())

	s.Require().NotPanics(func() {
		s.App.AppKeepers.CWHooksKeeper.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msg)
	})

	s.Require().Equal(callerLimit, ctx.GasMeter().GasConsumed())
	s.requireHookFailureEvent(ctx, contract)
}

// Hooks triggered by messages that a hook contract dispatches itself are not
// delivered.
func (s *KeeperTestSuite) TestNestedHookDeliveryIsSkipped() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(hookHeight)

	s.registerFailingHook(true, false)
	msg := s.validatorCreatedMsg()

	ctx := keeper.WithHookDispatch(s.Ctx).
		WithGasMeter(storetypes.NewGasMeter(1_000_000)).
		WithEventManager(sdk.NewEventManager())

	s.App.AppKeepers.CWHooksKeeper.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msg)

	s.Require().Zero(ctx.GasMeter().GasConsumed())
	s.Require().Empty(ctx.EventManager().Events())
}

// The number of contracts per hook type is bounded.
func (s *KeeperTestSuite) TestRegistrationIsCappedPerHook() {
	s.SetupTest()

	_, _, creator := testdata.KeyTestPubAddr()
	s.FundAcc(creator, sdk.NewCoins(sdk.NewCoin(s.bondDenom(), sdkmath.NewInt(1_000_000))))

	for range types.MaxContractsPerHook {
		contract := s.storeAndInstantiate(creator, failingContract)
		s.registerHook(contract, true, true)
	}

	extra := s.storeAndInstantiate(creator, failingContract)

	_, err := s.msgServer.RegisterStaking(s.Ctx, &types.MsgRegisterStaking{
		ContractAddress: extra.String(),
		RegisterAddress: s.authority(),
	})
	s.Require().ErrorContains(err, "at most")

	_, err = s.msgServer.RegisterGovernance(s.Ctx, &types.MsgRegisterGovernance{
		ContractAddress: extra.String(),
		RegisterAddress: s.authority(),
	})
	s.Require().ErrorContains(err, "at most")
}
