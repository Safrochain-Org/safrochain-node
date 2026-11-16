package types

// x/cw-hooks events.
const (
	// EventTypeContractExecutionFailed is emitted when a registered contract
	// fails to process a hook message (error, out of gas or panic). Hook
	// delivery is best-effort: the failure is isolated to that contract and
	// the staking or governance operation that triggered the hook proceeds.
	EventTypeContractExecutionFailed = "cw_hooks_contract_execution_failed"

	AttributeKeyHook     = "hook"
	AttributeKeyContract = "contract_address"
	AttributeKeyError    = "error"

	// HookStaking and HookGov are the values of AttributeKeyHook.
	HookStaking = "staking"
	HookGov     = "gov"
)
