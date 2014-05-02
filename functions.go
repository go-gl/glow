package main

import "sort"

type Function struct {
	Name       string
	GoName     string
	Parameters []Parameter
	Return     Type
	Doc        string
}

type Parameter struct {
	Name string
	Type Type
}

func (p Parameter) CName() string {
	return RenameIfReservedCWord(p.Name)
}

func (p Parameter) GoName() string {
	return RenameIfReservedGoWord(p.Name)
}

type Functions map[string]*Function
type SortedFunctions []*Function

func (sf SortedFunctions) Len() int           { return len(sf) }
func (sf SortedFunctions) Swap(i, j int)      { sf[i], sf[j] = sf[j], sf[i] }
func (sf SortedFunctions) Less(i, j int) bool { return sf[i].Name < sf[j].Name }

func (fs Functions) Sort() SortedFunctions {
	sortedFunctions := make(SortedFunctions, 0, len(fs))
	for _, f := range fs {
		sortedFunctions = append(sortedFunctions, f)
	}
	sort.Sort(sortedFunctions)
	return sortedFunctions
}
