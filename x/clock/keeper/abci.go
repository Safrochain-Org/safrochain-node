package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	helpers "github.com/Safrochain_Org/safrochain/app/helpers"
	"github.com/Safrochain_Org/safrochain/x/clock/types"
)

var endBlockSudoMessage = []byte(types.EndBlockSudoMessage)

// EndBlocker executes on contracts at the end of the block.
//
// Each unjailed contract runs in its own cached context with its own gas
// meter. A contract whose execution fails (error, out of gas or panic) has all
// of its state changes reverted and is jailed; the failure never affects the
// other contracts.
//
// The total gas consumed by clock contracts in a block is bounded by
// types.MaxBlockGas: a contract only starts if the remaining budget still
// covers a full ContractGasLimit. Iteration starts at an offset that rotates
// with the block height, so when the budget cannot cover every contract they
// are served round-robin and none is starved.
func EndBlocker(ctx context.Context, k Keeper) error {
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyEndBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(ctx)
	p := k.GetParams(ctx)

	contracts, err := k.GetAllContracts(ctx)
	if err != nil {
		logger.Error("Failed to get contracts", "error", err)
		return err
	}

	active := make([]string, 0, len(contracts))
	for _, contract := range contracts {
		if !contract.IsJailed {
			active = append(active, contract.ContractAddress)
		}
	}
	if len(active) == 0 {
		return nil
	}

	var (
		failed  []string
		gasUsed uint64
		offset  = int(uint64(sdkCtx.BlockHeight()) % uint64(len(active))) //nolint:gosec // modulo of a non-negative height
	)

	for i := range active {
		if types.MaxBlockGas-gasUsed < p.ContractGasLimit {
			logger.Debug("clock block gas budget exhausted", "executed", i, "registered", len(active))
			break
		}

		contractAddress := active[(offset+i)%len(active)]
		consumed, execErr := executeClockContract(sdkCtx, k, contractAddress, p.ContractGasLimit)
		gasUsed += consumed

		if execErr != nil {
			failed = append(failed, contractAddress)
			logger.Error("clock contract execution failed; jailing it", "contract", contractAddress, "error", execErr)

			if jailErr := k.SetJailStatus(ctx, contractAddress, true); jailErr != nil {
				logger.Error("Failed to jail contract", "contract", contractAddress, "error", jailErr)
			}
		}
	}

	if len(failed) > 0 {
		logger.Error("Failed to execute contracts", "contracts", failed)
	}

	return nil
}

// executeClockContract runs clock_end_block on a single contract in an
// isolated cached context and commits its writes and events only on success.
// wasmd's Sudo writes directly to the context it is given, so without the
// cache a contract that fails would keep the state it wrote before failing.
// It returns the gas the contract consumed (at most gasLimit).
func executeClockContract(ctx sdk.Context, k Keeper, contractAddress string, gasLimit uint64) (uint64, error) {
	contractAddr, err := sdk.AccAddressFromBech32(contractAddress)
	if err != nil {
		return 0, err
	}

	gasMeter := storetypes.NewGasMeter(gasLimit)
	cacheCtx, writeCache := ctx.CacheContext()
	cacheCtx = cacheCtx.WithGasMeter(gasMeter)

	var execErr error
	helpers.ExecuteContract(k.GetContractKeeper(), cacheCtx, contractAddr, endBlockSudoMessage, &execErr)
	if execErr != nil {
		return gasMeter.GasConsumedToLimit(), execErr
	}

	writeCache()
	return gasMeter.GasConsumedToLimit(), nil
}
