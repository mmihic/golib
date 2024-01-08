// Package ptr contains helpers to create pointers to various types.
package ptr

// To returns a pointer to the given type. Useful for initializing
// a reference from a primitive constant. e.g. a := ptr.To(100)
func To[T any](n T) *T {
	return &n
}

// Deref dereferences the given pointer, or returns
// a default value if the pointer is nil.
func Deref[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}

	return *ptr
}
