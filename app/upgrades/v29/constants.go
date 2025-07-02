package v29

import (
	"github.com/Safrochain_Org/safrochain/app/upgrades"
)

const UpgradeName = "v29"

const (
	expeditedMinDeposit = "10000000000"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV29UpgradeHandler,
}
