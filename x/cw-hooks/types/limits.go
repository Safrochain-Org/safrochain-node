package types

// MaxContractsPerHook bounds how many contracts can be registered for each hook
// type (staking, governance). Hooks are delivered on every delegation,
// validator-set change and governance action, and the gas they consume is
// charged to the operation that triggered them, so the registry must stay
// small.
const MaxContractsPerHook = 10
