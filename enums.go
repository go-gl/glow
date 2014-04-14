package main

import "sort"

type Enum struct {
	Name  string
	Value string
	Group string
}

type Enums map[string]*Enum
type SortedEnums []*Enum

func (f SortedEnums) Len() int           { return len(f) }
func (f SortedEnums) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f SortedEnums) Less(i, j int) bool { return f[i].Name < f[j].Name }

func (es Enums) Sort() SortedEnums {
	sortedEnums := make(SortedEnums, 0, len(es))
	for _, e := range es {
		sortedEnums = append(sortedEnums, e)
	}
	sort.Sort(sortedEnums)
	return sortedEnums
}
