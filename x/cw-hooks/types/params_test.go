package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Safrochain_Org/safrochain/x/cw-hooks/types"
)

func TestParamsValidate(t *testing.T) {
	for _, tc := range []struct {
		gasLimit uint64
		valid    bool
	}{
		{0, false},
		{types.MinContractGasLimit - 1, false},
		{types.MinContractGasLimit, true},
		{types.DefaultContractGasLimit, true},
		{types.MaxContractGasLimit, true},
		{types.MaxContractGasLimit + 1, false},
		{^uint64(0), false},
	} {
		err := types.NewParams(tc.gasLimit).Validate()
		if tc.valid {
			require.NoError(t, err, "gas limit %d", tc.gasLimit)
		} else {
			require.Error(t, err, "gas limit %d", tc.gasLimit)
		}
	}

	require.NoError(t, types.DefaultParams().Validate())
}
