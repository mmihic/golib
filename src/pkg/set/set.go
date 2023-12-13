// Package set contains a basic generic Set type.
package set

import "encoding/json"

// A Set is an unordered collection of values with the ability
// to test for inclusion and exclusion.
type Set[E comparable] map[E]struct{}

// NewSet creates a new set from a collection of values.
func NewSet[E comparable](vals ...E) Set[E] {
	set := make(Set[E], len(vals))
	for _, v := range vals {
		set.Add(v)
	}
	return set
}

// Add adds values to the set.
func (set Set[E]) Add(vals ...E) Set[E] {
	for _, v := range vals {
		set[v] = struct{}{}
	}
	return set
}

// Del removes values from the set.
func (set Set[E]) Del(vals ...E) Set[E] {
	for _, v := range vals {
		delete(set, v)
	}
	return set
}

// Has returns true if the set contains the given value.
func (set Set[E]) Has(val E) bool {
	_, ok := set[val]
	return ok
}

// All returns all values in the set.
func (set Set[E]) All() []E {
	result := make([]E, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}

// Len returns the size of the set.
func (set Set[E]) Len() int {
	return len(set)
}

// Intersect computes the intersection of two sets, producing
// a set that only contains the values present in both sets.
func (set Set[E]) Intersect(other Set[E]) Set[E] {
	result := NewSet[E]()
	for v := range set {
		if other.Has(v) {
			result.Add(v)
		}
	}

	return result
}

// Union computes the union of two sets, producing a set
// that contains the values present in either set.
func (set Set[E]) Union(other Set[E]) Set[E] {
	result := NewSet[E]()
	for v := range set {
		result.Add(v)
	}

	for v := range other {
		result.Add(v)
	}

	return result
}

// UnmarshalJSON unmarshals a set from a JSON array.
func (set *Set[E]) UnmarshalJSON(b []byte) error {
	var val []E
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}

	*set = NewSet(val...)
	return nil
}

// MarshalJSON marshals a set as a JSON array.
func (set Set[E]) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.All())
}
