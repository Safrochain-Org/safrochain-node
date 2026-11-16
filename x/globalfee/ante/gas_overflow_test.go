package ante

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A declared gas limit above MaxInt64 used to be cast to a negative int64,
// producing a negative required fee and a panic in sdk.NewCoin.
func TestGetMinGasPriceHugeGasDoesNotPanic(t *testing.T) {
	ctx := sdk.Context{}.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoinFromDec("usaf", sdkmath.LegacyNewDecWithPrec(5, 2))))

	for _, gas := range []uint64{0, 1, 200_000, math.MaxInt64, math.MaxInt64 + 1, math.MaxUint64} {
		require.NotPanics(t, func() {
			fees := GetMinGasPrice(ctx, gas)
			require.Len(t, fees, 1)
			require.False(t, fees[0].Amount.IsNegative(), "gas=%d", gas)
		}, "gas=%d", gas)
	}

	fees := GetMinGasPrice(ctx, math.MaxUint64)
	// ceil(0.05 * (2^64-1)) = 922337203685477581
	require.Equal(t, "922337203685477581", fees[0].Amount.String())
}
