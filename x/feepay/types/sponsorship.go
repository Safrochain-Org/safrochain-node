package types

// MaxSponsoredGas is the largest declared gas limit FeePay will sponsor for a
// single transaction. The sponsor is charged for the declared gas limit (not
// the gas actually used), so an unbounded limit would let any caller drain a
// contract's balance with one transaction and could overflow fee arithmetic.
const MaxSponsoredGas uint64 = 2_000_000
