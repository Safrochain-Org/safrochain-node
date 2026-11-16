package ante_test

import (
	"math"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Safrochain_Org/safrochain/testutil"
	"github.com/Safrochain_Org/safrochain/x/feepay/ante"
	feepaytypes "github.com/Safrochain_Org/safrochain/x/feepay/types"
	globalfeetypes "github.com/Safrochain_Org/safrochain/x/globalfee/types"
)

const (
	testDenom    = "usaf"
	testGasPrice = 1 // 1 usaf per gas unit, keeps the arithmetic obvious
)

type DeductFeeTestSuite struct {
	testutil.KeeperTestHelper

	decorator ante.DeductFeeDecorator
}

func TestDeductFeeTestSuite(t *testing.T) {
	suite.Run(t, new(DeductFeeTestSuite))
}

func (s *DeductFeeTestSuite) SetupTest() {
	s.Setup()

	k := s.App.AppKeepers
	s.decorator = ante.NewDeductFeeDecorator(
		k.FeePayKeeper,
		k.GlobalFeeKeeper,
		k.AccountKeeper,
		k.BankKeeper,
		k.FeeGrantKeeper,
		k.StakingKeeper,
	)

	s.Require().NoError(k.GlobalFeeKeeper.SetParams(s.Ctx, globalfeetypes.Params{
		MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(testDenom, sdkmath.LegacyNewDec(testGasPrice))),
	}))
}

// registerContract stores a FeePay contract directly (bypassing wasm) and
// funds the FeePay module account with the same amount.
func (s *DeductFeeTestSuite) registerContract(balance, walletLimit uint64) string {
	_, _, contract := testdata.KeyTestPubAddr()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
		ContractAddress: contract.String(),
		Balance:         balance,
		WalletLimit:     walletLimit,
	})
	if balance > 0 {
		s.FundModuleAcc(feepaytypes.ModuleName, sdk.NewCoins(sdk.NewCoin(testDenom, sdkmath.NewIntFromUint64(balance))))
	}
	return contract.String()
}

// feePayCtx returns a context flagged as being routed down the FeePay path.
func (s *DeductFeeTestSuite) feePayCtx() sdk.Context {
	state := feepaytypes.NewFeePayTxState()
	state.SetFeePayTx(true)
	return feepaytypes.WithFeePayTxState(s.Ctx, state)
}

func (s *DeductFeeTestSuite) newSender() sdk.AccAddress {
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin(testDenom, sdkmath.NewInt(1))))
	return sender
}

func executeMsg(sender sdk.AccAddress, contract string) *wasmtypes.MsgExecuteContract {
	return &wasmtypes.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: contract,
		Msg:      []byte("{}"),
	}
}

func (s *DeductFeeTestSuite) TestFeePayFailureDoesNotFallBackToFreeTx() {
	s.SetupTest()

	sender := s.newSender()
	// Contract is registered but cannot cover a 1000 gas tx.
	contract := s.registerContract(10, 5)

	tx := NewMockTx(sender, 1000, sdk.NewCoins(), executeMsg(sender, contract))
	_, err := s.decorator.AnteHandle(s.feePayCtx(), tx, false, noopAnte)
	s.Require().Error(err)
	s.Require().ErrorIs(err, feepaytypes.ErrContractNotEnoughFunds)
}

func (s *DeductFeeTestSuite) TestFeePaySuccess() {
	s.SetupTest()

	sender := s.newSender()
	contract := s.registerContract(10_000, 5)

	tx := NewMockTx(sender, 1000, sdk.NewCoins(), executeMsg(sender, contract))
	_, err := s.decorator.AnteHandle(s.feePayCtx(), tx, false, noopAnte)
	s.Require().NoError(err)

	fpc, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	s.Require().Equal(uint64(9_000), fpc.Balance)
}

// Usage limits are keyed on the contract caller (MsgExecuteContract sender),
// not on whoever pays the fee.
func (s *DeductFeeTestSuite) TestUsageKeyedOnContractCaller() {
	s.SetupTest()

	caller := s.newSender()
	payerA := s.newSender()
	payerB := s.newSender()
	contract := s.registerContract(10_000, 1)

	_, err := s.decorator.AnteHandle(s.feePayCtx(), NewMockTx(payerA, 1000, sdk.NewCoins(), executeMsg(caller, contract)), false, noopAnte)
	s.Require().NoError(err)

	// Same caller, different fee payer: the caller's limit (1) is exhausted.
	_, err = s.decorator.AnteHandle(s.feePayCtx(), NewMockTx(payerB, 1000, sdk.NewCoins(), executeMsg(caller, contract)), false, noopAnte)
	s.Require().ErrorIs(err, feepaytypes.ErrWalletExceededUsageLimit)

	fpc, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	uses, err := s.App.AppKeepers.FeePayKeeper.GetContractUses(s.Ctx, fpc, caller.String())
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), uses)
	uses, err = s.App.AppKeepers.FeePayKeeper.GetContractUses(s.Ctx, fpc, payerA.String())
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), uses)

	// A different caller is unaffected even with the same payer.
	other := s.newSender()
	_, err = s.decorator.AnteHandle(s.feePayCtx(), NewMockTx(payerA, 1000, sdk.NewCoins(), executeMsg(other, contract)), false, noopAnte)
	s.Require().NoError(err)
}

func (s *DeductFeeTestSuite) TestNonExecuteMsgIsRejectedWithoutPanic() {
	s.SetupTest()

	sender := s.newSender()
	msg := &banktypes.MsgSend{FromAddress: sender.String(), ToAddress: sender.String(), Amount: sdk.NewCoins(sdk.NewInt64Coin(testDenom, 1))}
	s.Require().NotPanics(func() {
		_, err := s.decorator.AnteHandle(s.feePayCtx(), NewMockTx(sender, 1000, sdk.NewCoins(), msg), false, noopAnte)
		s.Require().Error(err)
	})
}

func (s *DeductFeeTestSuite) TestSponsoredGasCap() {
	s.SetupTest()

	sender := s.newSender()
	contract := s.registerContract(10_000_000, 100)

	for _, gas := range []uint64{
		feepaytypes.MaxSponsoredGas + 1,
		math.MaxInt64,
		math.MaxInt64 + 1,
		math.MaxUint64,
	} {
		s.Require().NotPanics(func() {
			tx := NewMockTx(sender, gas, sdk.NewCoins(), executeMsg(sender, contract))
			_, err := s.decorator.AnteHandle(s.feePayCtx(), tx, false, noopAnte)
			s.Require().ErrorIs(err, feepaytypes.ErrSponsoredGasExceeded, "gas=%d", gas)
		}, "gas=%d", gas)
	}

	// Nothing was charged to the contract.
	fpc, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	s.Require().Equal(uint64(10_000_000), fpc.Balance)

	// Exactly at the cap is still sponsored.
	tx := NewMockTx(sender, feepaytypes.MaxSponsoredGas, sdk.NewCoins(), executeMsg(sender, contract))
	_, err = s.decorator.AnteHandle(s.feePayCtx(), tx, false, noopAnte)
	s.Require().NoError(err)
	fpc, err = s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	s.Require().Equal(10_000_000-feepaytypes.MaxSponsoredGas, fpc.Balance)
}

func (s *DeductFeeTestSuite) TestHugeGasLimitDoesNotPanicOnStandardPath() {
	s.SetupTest()

	sender := s.newSender()
	ctx := s.Ctx.WithIsCheckTx(true).WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoinFromDec(testDenom, sdkmath.LegacyNewDecWithPrec(5, 2))))

	for _, gas := range []uint64{math.MaxInt64 + 1, math.MaxUint64} {
		s.Require().NotPanics(func() {
			tx := NewMockTx(sender, gas, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 1)), executeMsg(sender, "x"))
			_, err := s.decorator.AnteHandle(ctx, tx, false, noopAnte)
			s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFee)
		}, "gas=%d", gas)
	}
}

func noopAnte(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

// MockTx is a minimal sdk.FeeTx.
type MockTx struct {
	feePayer   sdk.AccAddress
	feeGranter sdk.AccAddress
	gas        uint64
	fee        sdk.Coins
	msgs       []sdk.Msg
}

func NewMockTx(feePayer sdk.AccAddress, gas uint64, fee sdk.Coins, msgs ...sdk.Msg) MockTx {
	return MockTx{feePayer: feePayer, gas: gas, fee: fee, msgs: msgs}
}

func (tx MockTx) GetGas() uint64                     { return tx.gas }
func (tx MockTx) GetFee() sdk.Coins                  { return tx.fee }
func (tx MockTx) FeePayer() []byte                   { return tx.feePayer }
func (tx MockTx) FeeGranter() []byte                 { return tx.feeGranter }
func (tx MockTx) GetMsgs() []sdk.Msg                 { return tx.msgs }
func (MockTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (MockTx) ValidateBasic() error                  { return nil }
