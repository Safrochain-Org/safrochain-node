// Package v030 contains the v0.3.0 security upgrade.
//
// v0.3.0 is the first coordinated upgrade of the chain. It ships the
// SAF-AUDIT-2026-11 security fixes and the patched dependency set (wasmd
// v0.54.11 / wasmvm v2.2.10, cosmos-sdk v0.50.15, CometBFT v0.38.26, ibc-go
// v8.8.0). The handler only adjusts state that the new, stricter validation
// would reject or that the security fixes depend on; every step is
// idempotent and leaves already-valid values untouched.
package v030

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/Safrochain_Org/safrochain/app/upgrades"
)

// UpgradeName is the on-chain upgrade plan name.
const UpgradeName = "v0.3.0"

const (
	// BlockMaxGas replaces an unlimited (-1) block gas limit. Without a
	// limit a single block can carry an unbounded amount of computation.
	// 100M leaves ample room for the heaviest legitimate transactions
	// (wasm code uploads use a few million gas).
	BlockMaxGas int64 = 100_000_000

	// MinGasPrice is the network-wide minimum gas price, in the staking bond
	// denom, configured in x/globalfee when none is set. It matches the
	// default minimum-gas-prices written by `safrochaind init`. FeePay needs a
	// global price to compute what a sponsor pays.
	MinGasPrice = "0.05"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
