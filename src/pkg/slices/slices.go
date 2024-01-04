package slices

// Reverse reverses the given slice in place.
func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// Clone creates a clone of the given slice.
func Clone[S ~[]E, E any](s S) S {
	out := make(S, len(s))
	copy(out, s)
	return out
}

// Reversed returns a reversed copy of the given slice.
func Reversed[S ~[]E, E any](s S) S {
	out := Clone(s)
	Reverse(out)
	return out
}
