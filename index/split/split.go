package split

/*

this package implements a simple split index. "split" because we store two mini-indexes
and join them at search time using the join key for that index.

join keys can be anything you use to associate some metrics with some tags. Our first
common case was hostname: associating metrics sent by a host with
properties of that host (liveness, dc, rack ID, etc.)

the two pieces of the index look like this (using hostname as an example):

"left" side of the join (tags to $join_key, in this case hostnames)
left: {
	server-state:live: [hostname-1234, hostname-1235, ...]
}

"right" side of the join ($join_key to metrics)
right: {
	hostname-1234: ["server.hostname-1234.cpu.i7z", ...]
}

the query process goes like this:

1) user searches for metrics that match a set of tags ("server-state:live", "server-dc:lhr")
2) we reach into the "left" index to fetch all of the join keys (hostnames) associated with all of these tags
3) take that intersection of join keys to find all the metrics associated with them
4) success! return that set of metrics

*/

import (
	"fmt"
	"sync"
)

type Index struct {
	joinKey string

	tagToJoin map[string]map[string]bool
	tagMutex  sync.RWMutex
	tagCount  int

	joinToMetric map[string]map[string]bool
	metricMutex  sync.RWMutex
	metricCount  int
}

func NewIndex(joinKey string) *Index {
	return &Index{
		joinKey: joinKey,

		tagToJoin: make(map[string]map[string]bool),

		joinToMetric: make(map[string]map[string]bool),
	}
}

func (i *Index) AddMetrics(join string, metrics []string) error {
	if len(metrics) == 0 {
		return fmt.Errorf("cannot add 0 metrics to join %q", join)
	}

	i.metricMutex.Lock()
	defer i.metricMutex.Unlock()

	metricIndex, ok := i.joinToMetric[join]
	if !ok {
		metricIndex = make(map[string]bool)
		i.joinToMetric[join] = metricIndex
	}

	for _, metric := range metrics {
		_, ok := metricIndex[metric]
		// this only needs to branch in order to avoid double-counting things
		if !ok {
			i.metricCount++
			metricIndex[metric] = true
		}
	}

	return nil
}

func (i *Index) AddTags(join string, tags []string) error {
	if len(tags) == 0 {
		return fmt.Errorf("cannot add 0 tags to join %q", join)
	}

	i.tagMutex.Lock()
	defer i.tagMutex.Unlock()

	for _, tag := range tags {
		tagIndex, ok := i.tagToJoin[tag]
		if !ok {
			tagIndex = make(map[string]bool)
			i.tagToJoin[tag] = tagIndex
		}
		_, ok = tagIndex[join]
		// this only needs to branch in order to avoid double-counting things
		if !ok {
			i.tagCount++
			tagIndex[join] = true
		}
	}

	return nil
}

func (i *Index) Query(query []string) ([]string, error) {
	joinKeyCounts := make(map[string]int)
	// get a slice of all the join keys (for example, hostnames) associated with these tags
	i.tagMutex.RLock()
	for _, tag := range query {
		for key := range i.tagToJoin[tag] {
			joinKeyCounts[key]++
		}
	}
	i.tagMutex.RUnlock()

	// pluck out the join keys present for every search tag (intersection)
	var joinKeys []string
	for key, count := range joinKeyCounts {
		if count == len(query) {
			joinKeys = append(joinKeys, key)
		}
	}

	var metrics []string
	// now use those join keys to fetch out metrics
	i.metricMutex.RLock()
	for _, joinKey := range joinKeys {
		for metric := range i.joinToMetric[joinKey] {
			metrics = append(metrics, metric)
		}
	}
	i.metricMutex.RUnlock()
	return metrics, nil
}

func (i *Index) Name() string {
	return i.joinKey
}

func (i *Index) TagSize() int {
	// or convert the sizes to atomics
	i.tagMutex.RLock()
	defer i.tagMutex.RUnlock()
	return i.tagCount
}

func (i *Index) MetricSize() int {
	// or convert the sizes to atomics
	i.metricMutex.RLock()
	defer i.metricMutex.RUnlock()
	return i.metricCount
}