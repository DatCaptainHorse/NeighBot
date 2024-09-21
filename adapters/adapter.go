package adapters

import (
	"NeighBot/llm"
)

type ChatAdapterConfig struct {
	Enabled     bool              `json:"enabled"`
	MemoryStore *llm.MemoryStore  `json:"-"`
	LLMClient   *llm.OpenAIClient `json:"-"`
}

// ChatAdapter is a common interface all adapters must implement
type ChatAdapter interface {
	SetConfig(cfg interface{}) error
	Initialize() error
	Start() error
	Stop() error
	AdapterName() string
}
