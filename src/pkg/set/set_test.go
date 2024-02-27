package set

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, []string{"a", "b", "c", "d", "f", "g"}, all)

	union = set2.Union(set1)
	all = union.All()
	sort.Strings(all)
	assert.Equal(t, []string{"a", "b", "c", "d", "f", "g"}, all)
}

func TestSet_Contains(t *testing.T) {
	set1 := NewSet("a", "b", "c", "d")
	assert.True(t, set1.Contains(NewSet("a", "b")))
	assert.True(t, set1.Contains(NewSet("c", "b")))
	assert.False(t, set1.Contains(NewSet("a", "b", "c", "d", "e")))
	assert.False(t, set1.Contains(NewSet("a", "g")))
	assert.False(t, set1.Contains(NewSet("z")))
}

func TestSet_Equal(t *testing.T) {
	set1 := NewSet("a", "b", "c", "d")
	assert.True(t, set1.Equal(NewSet("a", "b", "c", "d")))
	assert.False(t, set1.Equal(NewSet("a", "b", "c", "d", "e")))
	assert.False(t, set1.Equal(NewSet("c", "b")))
	assert.False(t, set1.Equal(NewSet("a", "b", "c", "g")))
}

func TestSet_MarshalJSON(t *testing.T) {
	set := NewSet("a", "b", "c", "d")
	output, err := json.Marshal(set)
	require.NoError(t, err)

	// should be able to unmarshal into an array
	var values []string
	err = json.Unmarshal(output, &values)
	require.NoError(t, err)
	sort.Strings(values)
	assert.Equal(t, values, []string{"a", "b", "c", "d"})
}

func TestSet_UnmarshalJSON(t *testing.T) {
	var set Set[string]
	err := json.Unmarshal([]byte(`["a", "b", "c", "d"]`), &set)
	require.NoError(t, err)

	all := set.All()
	sort.Strings(all)
	assert.Equal(t, []string{"a", "b", "c", "d"}, all)
}
