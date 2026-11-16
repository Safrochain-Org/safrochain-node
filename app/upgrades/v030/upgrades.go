package v030

import (
	"context"
	"errors"
	"fmt"
	"slices"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Safrochain_Org/safrochain/app/keepers"
	clocktypes "github.com/Safrochain_Org/safrochain/x/clock/types"
	cwhookstypes "github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

// step is one idempotent state adjustment performed by the upgrade.
type step struct {
	name string
	run  func(ctx sdk.Context, k *keepers.AppKeepers, logger log.Logger) error
}

var steps = []step{
	{"consensus block max gas", setBlockMaxGas},
	{"x/clock params", sanitizeClockParams},
	{"x/cw-hooks params", sanitizeCWHooksParams},
	{"x/drip params", sanitizeDripParams},
	{"x/globalfee minimum gas prices", setMinimumGasPrices},
	{"x/tokenfactory params", validateTokenFactoryParams},
	{"x/feeshare params", validateFeeShareParams},
	{"ica host allow messages", restrictICAHostMessages},
}

// CreateUpgradeHandler returns the v0.3.0 upgrade handler.
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrade", UpgradeName)

		logger.Info("running module migrations", "from", vm)
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}

		for _, s := range steps {
			if err := s.run(sdkCtx, k, logger); err != nil {
				return nil, fmt.Errorf("%s: %w", s.name, err)
			}
		}

		logger.Info("upgrade complete", "to", versionMap)
		return versionMap, nil
	}
}

// setBlockMaxGas bounds the block gas limit when it is unlimited.
func setBlockMaxGas(ctx sdk.Context, k *keepers.AppKeepers, logger log.Logger) error {
	cp, err := k.ConsensusParamsKeeper.ParamsStore.Get(ctx)
	if err != nil {
		return err
	}
	if cp.Block == nil {
		return errors.New("consensus block params are not set")
	}
	if cp.Block.MaxGas > 0 {
		return nil
	}

	logger.Info("setting block max gas", "from", cp.Block.MaxGas, "to", BlockMaxGas)
	cp.Block.MaxGas = BlockMaxGas
	return k.ConsensusParamsKeeper.ParamsStore.Set(ctx, cp)
}

// sanitizeClockParams resets a contract gas limit outside the bounds enforced
// since SAF-03 (the live testnet has 100,000, below the 1,000,000 minimum).
func sanitizeClockParams(ctx sdk.Context, k *keepers.AppKeepers, logger log.Logger) error {
	p := k.ClockKeeper.GetParams(ctx)
	if p.Validate() == nil {
		return nil
	}

	logger.Info("resetting x/clock contract gas limit", "from", p.ContractGasLimit, "to", clocktypes.DefaultContractGasLimit)
	p.ContractGasLimit = clocktypes.DefaultContractGasLimit
	return k.ClockKeeper.SetParams(ctx, p)
}

// sanitizeCWHooksParams resets a contract gas limit outside the new bounds.
func sanitizeCWHooksParams(ctx sdk.Context, k *keepers.AppKeepers, logger log.Logger) error {
	p := k.CWHooksKeeper.GetParams(ctx)
	if p.Validate() == nil {
		return nil
	}

	logger.Info("resetting x/cw-hooks contract gas limit", "from", p.ContractGasLimit, "to", cwhookstypes.DefaultContractGasLimit)
	p.ContractGasLimit = cwhookstypes.DefaultContractGasLimit
	return k.CWHooksKeeper.SetParams(ctx, p)
}

// sanitizeDripParams disables drip when it is enabled with an empty allow-list
// (rejected since SAF-06). With an empty allow-list nobody can drip, so this
// does not change behaviour.
func sanitizeDripParams(ctx sdk.Context, k *keepers.AppKeepers, logger log.Logger) error {
	p := k.DripKeeper.GetParams(ctx)
	if p.EnableDrip && len(p.AllowedAddresses) == 0 {
		logger.Info("disabling x/drip: enabled with an empty allow-list")
		p.EnableDrip = false
	}

	if err := p.Validate(); err != nil {
		return err
	}
	return k.DripKeeper.SetParams(ctx, p)
}

// setMinimumGasPrices sets the network-wide minimum gas price when none is
// configured.
func setMinimumGasPrices(ctx sdk.Context, k *keepers.AppKeepers, logger log.Logger) error {
	p := k.GlobalFeeKeeper.GetParams(ctx)
	if len(p.MinimumGasPrices) > 0 {
		return p.Validate()
	}

	bondDenom, err := k.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	price, err := sdkmath.LegacyNewDecFromStr(MinGasPrice)
	if err != nil {
		return err
	}

	p.MinimumGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(bondDenom, price))
	logger.Info("setting x/globalfee minimum gas prices", "to", p.MinimumGasPrices.String())

	if err := p.Validate(); err != nil {
		return err
	}
	return k.GlobalFeeKeeper.SetParams(ctx, p)
}

// validateTokenFactoryParams fails the upgrade if the stored params do not
// pass validation, instead of silently rewriting the denom creation fee.
func validateTokenFactoryParams(ctx sdk.Context, k *keepers.AppKeepers, _ log.Logger) error {
	return k.TokenFactoryKeeper.GetParams(ctx).Validate()
}

// validateFeeShareParams fails the upgrade if the stored params do not pass
// validation.
func validateFeeShareParams(ctx sdk.Context, k *keepers.AppKeepers, _ log.Logger) error {
	return k.FeeShareKeeper.GetParams(ctx).Validate()
}

// restrictICAHostMessages replaces the "*" wildcard (the ibc-go default) with
// the least-privilege allow-list. The live network has no ICA channels.
func restrictICAHostMessages(ctx sdk.Context, k *keepers.AppKeepers, logger log.Logger) error {
	p := k.ICAHostKeeper.GetParams(ctx)
	if !slices.Contains(p.AllowMessages, icahosttypes.AllowAllHostMsgs) {
		return nil
	}

	logger.Info("restricting ICA host allow messages", "count", len(keepers.ICAHostAllowMessages))
	p.AllowMessages = slices.Clone(keepers.ICAHostAllowMessages)
	if err := p.Validate(); err != nil {
		return err
	}

	k.ICAHostKeeper.SetParams(ctx, p)
	return nil
}
