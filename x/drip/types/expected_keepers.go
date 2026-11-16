package types

import "context"

// StakingKeeper defines the expected staking keeper used to resolve the native bond denom.
type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}
