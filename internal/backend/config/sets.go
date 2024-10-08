package config

import (
	"reflect"
	"slices"
)

type set struct {
	name   string
	set    baseSet
	chosen ComponentConfig
}

var sets = make(map[string]set)

func RegisterSet[T ComponentConfig](name string, givenSet Set[T], chosen T) {
	sets[name] = set{
		name:   name,
		set:    givenSet,
		chosen: chosen,
	}
}

func ListAll() []string {
	var all []string
	for name, set := range sets {
		all = append(all, parseKoanfTags(name, reflect.Indirect(reflect.ValueOf(set.chosen)))...)
	}

	slices.Sort(all)

	return all
}
