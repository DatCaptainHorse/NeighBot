package config

import (
	"NeighBot/adapters"
	"encoding/json"
	"os"
)

type MainConfig struct {
	Adapters AdaptersConfig `json:"adapters"`
	LLM      LLMConfig      `json:"llm"`
}

type AdaptersConfig struct {
	Configs map[string]interface{} `json:"configs"`
}

type LLMConfig struct {
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
}

func (cfg *MainConfig) Load(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, cfg); err != nil {
		return err
	}

	if cfg.Adapters.Configs == nil {
		cfg.Adapters.Configs = make(map[string]interface{})
	}
	for _, adapterName := range adapters.RegisteredAdapters() {
		if _, exists := cfg.Adapters.Configs[adapterName]; !exists {
			cfg.Adapters.Configs[adapterName] = adapters.NewAdapterDefaultConfig(adapterName)
		}
	}

	return nil
}

func (cfg *MainConfig) Save(configPath string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (cfg *MainConfig) CreateDefault(configPath string) error {
	cfg.LLM = LLMConfig{
		APIKey:   "-",
		Endpoint: "http://localhost:8000",
		Model:    "my-default-model",
	}

	cfg.Adapters.Configs = make(map[string]interface{})
	for _, adapterName := range adapters.RegisteredAdapters() {
		cfg.Adapters.Configs[adapterName] = adapters.NewAdapterDefaultConfig(adapterName)
	}

	return cfg.Save(configPath)
}
