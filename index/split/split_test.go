package split

import (
	"testing"

	"github.com/kanatohodets/carbonsearch/index"
	"github.com/kanatohodets/carbonsearch/util/test"
)

func TestSortJoins(t *testing.T) {
	// make sure it doesn't error on a 0 item slice
	joins := []Join{}
	SortJoins(joins)

	// 1 item
	joins = []Join{HashJoin("foo")}
	expectedFirst := joins[0]
	SortJoins(joins)
	if joins[0] != expectedFirst || len(joins) > 1 {
		t.Errorf("index test: SortJoins wrecked a 1 item slice, somehow")
		return
	}

	// create a deliberately unsorted 2 item list
	joins = []Join{
		HashJoin("foo"),
		HashJoin("bar"),
	}
	a, b := joins[0], joins[1]
	expectedFirst = a
	if b > a {
		joins = []Join{b, a}
	} else {
		expectedFirst = b
	}

	SortJoins(joins)
	if joins[0] != expectedFirst {
		t.Errorf("index test: SortJoins did not sort the slice: expected %v as first item, but got %v", expectedFirst, joins[0])
	}

}

func TestQuery(t *testing.T) {
	metricName := "server.hostname-1234"
	host := "hostname-1234"

	in := NewIndex("host")
	metrics := index.HashMetrics([]string{metricName})
	tags := index.HashTags([]string{"server-state:live", "server-dc:lhr"})
	query := index.NewQuery([]string{"server-state:live"})

	in.AddMetrics(host, metrics)
	in.AddTags(host, tags)
	result, err := in.Query(query)
	if err != nil {
		t.Error(err)
	}

	if len(result) == 1 {
		if result[0] != index.HashMetric(metricName) {
			t.Errorf("split index test: %v was not found in the index", metricName)
		}
	} else {
		t.Errorf("split index test: the index had %d search results. that value is wrong because it isn't 1", len(result))
	}

	emptyResult, err := in.Query(index.NewQuery([]string{"blorgtag"}))
	if err != nil {
		t.Errorf("error querying blorgtag: %v", err)
	}
	if len(emptyResult) != 0 {
		t.Errorf("split index test: found some results on a bogus query: %v", emptyResult)
	}
}

func BenchmarkSmallsetQuery(b *testing.B) {
	metricName := "server.hostname-1234"
	host := "hostname-1234"
	in := NewIndex("host")
	tags := []string{"server-state:live", "server-dc:lhr"}
	metrics := index.HashMetrics([]string{metricName})

	in.AddMetrics(host, metrics)
	in.AddTags(host, index.HashTags(tags))

	query := index.NewQuery([]string{"server-state:live"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		in.Query(query)
	}
}

func BenchmarkLargesetQuery(b *testing.B) {
	b.StopTimer()
	in := NewIndex("host")
	hosts := test.GetJoinCorpus(100)
	queryTerms := []string{}
	for _, host := range hosts {
		in.AddMetrics(host, index.HashMetrics(test.GetMetricCorpus(1000)))
		tags := test.GetTagCorpus(10)
		if test.Rand().Intn(15) == 1 {
			queryTerms = append(queryTerms, tags[test.Rand().Int()%len(tags)])
		}
		in.AddTags(host, index.HashTags(tags))
	}

	query := index.NewQuery(queryTerms)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		in.Query(query)
	}
}

// TODO(btyler) consolidate this into a testing table
func TestIntersectJoins(t *testing.T) {
	// basic intersection
	joins := [][]Join{
		HashJoins([]string{"foo", "bar", "baz"}),
		HashJoins([]string{"qux", "bar"}),
		HashJoins([]string{"blorg", "bar"}),
	}

	for _, joinList := range joins {
		SortJoins(joinList)
	}

	expectedList := HashJoins([]string{"bar"})
	expected := map[Join]bool{}

	for _, join := range expectedList {
		expected[join] = false
	}

	intersection := IntersectJoins(joins)

	for _, join := range intersection {
		_, ok := expected[join]
		if !ok {
			t.Errorf("index test: join intersect included %v, which was not expected", join)
			return
		}
		expected[join] = true
	}

	for join, found := range expected {
		if !found {
			t.Errorf("index test: join intersect did NOT include %v, which was expected to be there", join)
		}
	}

	// empty intersection due to empty universe
	intersection = IntersectJoins([][]Join{})
	if len(intersection) > 0 {
		t.Error("index test: join intersect on empty set returned non-empty")
	}

	// empty intersection due to one empty subset
	joins = [][]Join{
		HashJoins([]string{"foo", "bar", "baz"}),
		HashJoins([]string{"qux", "bar"}),
		HashJoins([]string{}),
	}

	for _, joinList := range joins {
		SortJoins(joinList)
	}
	intersection = IntersectJoins(joins)
	if len(intersection) > 0 {
		t.Error("index test: join intersect returned non-empty, but it should have been empty")
	}

	// empty intersection due to no overlap
	joins = [][]Join{
		HashJoins([]string{"foo"}),
		HashJoins([]string{"bar"}),
		HashJoins([]string{"baz", "blorg", "buzz", "pow", "kablooie", "whizbang", "rain", "always rain"}),
	}
	for _, joinList := range joins {
		SortJoins(joinList)
	}
	intersection = IntersectJoins(joins)
	if len(intersection) > 0 {
		t.Error("index test: join intersect returned non-empty, but it should have been empty")
	}

	// intersection of just one thing
	joins = [][]Join{HashJoins([]string{"foo"})}
	for _, joinList := range joins {
		SortJoins(joinList)
	}
	intersection = IntersectJoins(joins)
	if len(intersection) != 1 {
		t.Error("index test: join intersect returned more than 1 result for a universe of 1")
		return
	}
	if intersection[0] != joins[0][0] {
		t.Error("index test: somehow a universe of 1 resulted in an intersection of 1, but not that 1. wtf o_o")
		return
	}
}
