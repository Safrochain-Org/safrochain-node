package keepers

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

const (
	// DefaultSafrochainInstanceCost is initially set the same as in wasmd
	DefaultSafrochainInstanceCost uint64 = 60_000
	// DefaultSafrochainCompileCost set to a large number for testing
	DefaultSafrochainCompileCost uint64 = 3
)

// SafrochainGasRegisterConfig is defaults plus a custom compile amount
func SafrochainGasRegisterConfig() wasmtypes.WasmGasRegisterConfig {
	gasConfig := wasmtypes.DefaultGasRegisterConfig()
	gasConfig.InstanceCost = DefaultSafrochainInstanceCost
	gasConfig.CompileCost = DefaultSafrochainCompileCost

	return gasConfig
}

func NewSafrochainWasmGasRegister() wasmtypes.WasmGasRegister {
	return wasmtypes.NewWasmGasRegister(SafrochainGasRegisterConfig())
}
