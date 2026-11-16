package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// WithHookDispatch marks ctx as derived from a hook delivery (test helper).
func WithHookDispatch(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(hookDispatchKey{}, true)
}
