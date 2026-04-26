package v28

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/Safrochain_Org/safrochain/app/upgrades"
)

const UpgradeName = "v28"

// mevModuleAccount is the bech32 address of the legacy skip-mev/pob "builder"
// module account. v28 sweeps any residual MEV profits from this account into
// the community pool before the module store is deleted (see StoreUpgrades
// below). The bytes are derived from Cosmos SDK's address.Module("builder")
// (i.e. sha256("module" || "builder")[:20]) and re-encoded with the chain's
// addr_safro account prefix; the previously-hardcoded "safrochain1..."
// string had an invalid bech32 checksum and would have caused
// MustAccAddressFromBech32 to panic mid-upgrade.
const (
	mevModuleAccount = "addr_safro149y9yqeqmn66hpv0ydjtknpr9tmympyzyl854m"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV28UpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{
			"08-wasm",
			"builder",
		},
	},
}
