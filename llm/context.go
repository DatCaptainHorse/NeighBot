package llm

import (
	"NeighBot/filters"
	"NeighBot/logger"
)

type StoredContext struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Messages        []StoredMessage        `json:"-"`
	Filters         map[string]bool        `json:"filters"`
	FilterManager   *filters.FilterManager `json:"-"`
	AssociatedChats []string               `json:"associated_chats"` // TODO: For now it's discord channel IDs
}

func (ctx *StoredContext) AddMessage(message StoredMessage) {
	ctx.Messages = append(ctx.Messages, message)
}

func (ctx *StoredContext) ApplyFilters(input string) string {
	if ctx.FilterManager == nil {
		ctx.InitializeFilters()
	}
	return ctx.FilterManager.Apply(input)
}

func (ctx *StoredContext) InitializeFilters() {
	ctx.FilterManager = &filters.FilterManager{}
	for filterName, enabled := range ctx.Filters {
		if !enabled {
			continue
		}

		filter, exists := filters.GetFilter(filterName)
		if !exists {
			logger.Sugar.Warnw("Unknown filter", "filter_name", filterName, "context_id", ctx.ID)
			continue
		}

		ctx.FilterManager.AddFilter(filter)
	}
}

func (ctx *StoredContext) SerializeConfig() map[string]interface{} {
	return map[string]interface{}{
		"context_id":  ctx.ID,
		"name":        ctx.Name,
		"description": ctx.Description,
	}
}

func (ctx *StoredContext) DeserializeConfig(data map[string]interface{}) {
	if val, ok := data["context_id"].(string); ok {
		ctx.ID = val
	}
	if val, ok := data["name"].(string); ok {
		ctx.Name = val
	}
	if val, ok := data["description"].(string); ok {
		ctx.Description = val
	}
}
