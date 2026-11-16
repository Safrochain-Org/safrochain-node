package types

import "context"

// StakingKeeper defines the subset of the x/staking keeper used by FeePay to
// resolve the chain's bond denom at runtime.
type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}
