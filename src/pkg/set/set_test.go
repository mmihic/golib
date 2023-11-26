package set

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	set := NewSet("a", "b", "c", "d")
	assert.True(t, set.Has("a"))
	assert.True(t, set.Has("d"))
	assert.False(t, set.Has("f"))

	all := set.All()
	sort.Strings(all)
	assert.Equal(t, []string{"a", "b", "c", "d"}, all)

	set.Add("f").Add("g").Del("a")
	assert.True(t, set.Has("f"))
	assert.True(t, set.Has("g"))
	assert.False(t, set.Has("a"))

	all = set.All()
	sort.Strings(all)
	assert.Equal(t, []string{"b", "c", "d", "f", "g"}, all)
}

func TestSet_Intersect(t *testing.T) {
	set1 := NewSet("a", "b", "c", "d")
	set2 := NewSet("a", "c", "g", "f")

	intersect := set1.Intersect(set2)
	all := intersect.All()
	sort.Strings(all)
	assert.Equal(t, []string{"a", "c"}, all)

	intersect = set2.Intersect(set1)
	all = intersect.All()
	sort.Strings(all)
	assert.Equal(t, []string{"a", "c"}, all)
}

func TestSet_Union(t *testing.T) {
	set1 := NewSet("a", "b", "c", "d")
	set2 := NewSet("a", "c", "g", "f")

	union := set1.Union(set2)
	all := union.All()
	sort.Strings(all)
	assert.Equal(t, []string{"a", "b", "c", "d", "g", "f"}, all)

	union = set2.Union(set1)
	all = union.All()
	sort.Strings(all)
	assert.Equal(t, []string{"a", "b", "c", "d", "g", "f"}, all)
}
