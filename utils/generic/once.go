package generic

import "sync"

// Once returns a function that calls f once and then returns its result.
//
// Deprecated: use sync.OnceValue.
func Once[T any](f func() T) func() T {
	return sync.OnceValue(f)
}
