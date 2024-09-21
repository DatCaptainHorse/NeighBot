package filters

import "NeighBot/logger"

var filterRegistry = map[string]Filter{}

func RegisterFilter(filter Filter) {
	filterRegistry[filter.Name()] = filter
	logger.Sugar.Infow("Filter registered", "filter_name", filter.Name())
}

func GetFilter(name string) (Filter, bool) {
	filter, exists := filterRegistry[name]
	if !exists {
		logger.Sugar.Warnw("Filter not found", "filter_name", name)
	}
	return filter, exists
}

func InitializeFilters() {
	RegisterFilter(EmojiFilter{})
	RegisterFilter(EmphasisFilter{})
	RegisterFilter(LinkFilter{})
	logger.Sugar.Info("Filters initialized")
}
