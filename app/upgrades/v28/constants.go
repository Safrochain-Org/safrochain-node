package v28

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/Safrochain_Org/safrochain/app/upgrades"
)

const UpgradeName = "v28"

const (
	mevModuleAccount = "safrochain1ma4sw9m2nvtucny6lsjhh4qywvh86zdh5dlkd4"
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
