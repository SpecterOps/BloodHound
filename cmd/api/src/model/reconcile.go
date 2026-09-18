package model

// ReconcileResult represents the persisted outcome of reconciling a desired input collection against existing records.
// Slice ordering is not part of the contract.
type ReconcileResult[T any] struct {
	Created []T
	Updated []T
	Deleted []T
}
