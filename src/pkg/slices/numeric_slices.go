package slices

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// Median finds the median of a slice of numeric values.
func Median[T constraints.Integer | constraints.Float, S ~[]T](s S) T {
	nCopy := slices.Clone(s)
	slices.Sort(nCopy)
	ln := len(nCopy)
	switch {
	case ln == 0:
		return 0
	case ln%2 == 0:
		n, m := nCopy[ln/2-1], nCopy[ln/2]
		return (n + m) / T(2)
	default:
		return nCopy[ln/2]
	}
}

// Max finds the max value in a slice of numeric values.
func Max[T constraints.Integer | constraints.Float, S ~[]T](s S) T {
	var nMax T
	for i, n := range s {
		if i == 0 {
			nMax = n
		} else if nMax < n {
			nMax = n
		}
	}

	return nMax
}

// Min finds the min value in a slice of numeric values.
func Min[T constraints.Integer | constraints.Float, S ~[]T](s S) T {
	var nMin T
	for i, n := range s {
		if i == 0 {
			nMin = n
		} else if nMin > n {
			nMin = n
		}
	}

	return nMin
}

// Mean finds the average in a slice of numeric values.
func Mean[T constraints.Integer | constraints.Float, S ~[]T](s S) T {
	var nMean T
	for i, n := range s {
		nMean = ((nMean * T(i)) + n) / T(i+1)
	}
	return nMean
}
