package filters

type Filter interface {
	Apply(input string) string
	Name() string
}

type FilterManager struct {
	filters []Filter
}

func (fm *FilterManager) AddFilter(filter Filter) {
	fm.filters = append(fm.filters, filter)
}

func (fm *FilterManager) Apply(input string) string {
	for _, filter := range fm.filters {
		input = filter.Apply(input)
	}
	return input
}
