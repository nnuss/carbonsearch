package util

import (
	"expvar"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
)

type Stats struct {
	TagMessages *expvar.Int
	TagsIndexed *expvar.Int

	MetricMessages *expvar.Int
	MetricsIndexed *expvar.Int

	CustomMessages   *expvar.Int
	FullIndexTags    *expvar.Int
	FullIndexMetrics *expvar.Int

	QueriesHandled     *expvar.Int
	QueryTagsByService *expvar.Map

	ServicesByIndex *expvar.Map

	SplitIndexes *expvar.Map
}

func InitStats() *Stats {
	return &Stats{
		TagMessages: expvar.NewInt("TagMessages"),
		TagsIndexed: expvar.NewInt("TagsIndexed"),

		MetricMessages: expvar.NewInt("MetricMessages"),
		MetricsIndexed: expvar.NewInt("MetricsIndexed"),

		CustomMessages:   expvar.NewInt("CustomMessages"),
		FullIndexTags:    expvar.NewInt("FullIndexTags"),
		FullIndexMetrics: expvar.NewInt("FullIndexMetrics"),

		QueriesHandled:     expvar.NewInt("QueriesHandled"),
		QueryTagsByService: expvar.NewMap("QueryTagsByService"),

		SplitIndexes: expvar.NewMap("SplitIndexes"),

		ServicesByIndex: expvar.NewMap("ServicesByIndex"),
	}
}

type ExpInt int

func (i ExpInt) String() string { return strconv.Itoa(int(i)) }

type ExpString string

func (s ExpString) String() string { return string(s) }

func ReadConfig(path string, dest interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error while reading path %q: %s", path, err)
	}

	err = yaml.Unmarshal(bytes, dest)
	if err != nil {
		return fmt.Errorf("error parsing %q: %s", path, err)
	}
	return nil
}