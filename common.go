package main

import (
	"NeighBot/adapters"
	"NeighBot/adapters/discord"
	"NeighBot/config"
	"NeighBot/llm"
	"NeighBot/logger"
	"encoding/json"
	"fmt"
	"reflect"
)

var runningAdapters []adapters.ChatAdapter

func HandleAdapters(cfg *config.MainConfig, memoryStore *llm.MemoryStore, llmClient *llm.OpenAIClient) error {
	/* Adapter register list */
	if err := adapters.RegisterAdapter("discord", &discord.DiscordAdapter{}, discord.DiscordConfig{}); err != nil {
		return err
	}
	/* End of adapter register list */

	// Ensure all registered adapters have a config entry
	for _, adapterName := range adapters.RegisteredAdapters() {
		if _, exists := cfg.Adapters.Configs[adapterName]; !exists {
			cfg.Adapters.Configs[adapterName] = adapters.NewAdapterDefaultConfig(adapterName)
		}
	}

	for adapterName, adapterConfigData := range cfg.Adapters.Configs {
		adapter, configType, err := adapters.CreateAdapter(adapterName)
		if err != nil {
			return fmt.Errorf("create adapter %s: %w", adapterName, err)
		}

		adapterConfig := reflect.New(configType).Interface()
		configVal, err := json.Marshal(adapterConfigData)
		if err != nil {
			return fmt.Errorf("marshal adapter config %s: %w", adapterName, err)
		}
		if err = json.Unmarshal(configVal, adapterConfig); err != nil {
			return fmt.Errorf("unmarshal adapter config %s: %w", adapterName, err)
		}

		// Set shared dependencies
		baseConfig := extractBaseConfig(adapterConfig)
		if baseConfig == nil || !baseConfig.Enabled {
			logger.Sugar.Infow("Adapter disabled", "adapter", adapterName)
			continue
		}
		baseConfig.MemoryStore = memoryStore
		baseConfig.LLMClient = llmClient

		// Pass the config to the adapter
		if err = adapter.SetConfig(adapterConfig); err != nil {
			return fmt.Errorf("set config for adapter %s: %w", adapterName, err)
		}

		// Initialize and start the adapter
		if err = adapter.Initialize(); err != nil {
			return fmt.Errorf("initialize adapter %s: %w", adapterName, err)
		}
		if err = adapter.Start(); err != nil {
			logger.Sugar.Errorw("Failed to start adapter", "adapter_name", adapterName, "error", err)
			continue
		}

		logger.Sugar.Infow("Adapter started", "adapter", adapterName)
		runningAdapters = append(runningAdapters, adapter)
	}

	return nil
}

func StopAdapters() {
	for _, adapter := range runningAdapters {
		logger.Sugar.Infow("Stopping adapter", "adapter_name", adapter.AdapterName())
		if err := adapter.Stop(); err != nil {
			logger.Sugar.Errorw("Failed to stop adapter", "adapter_name", adapter.AdapterName(), "error", err)
		} else {
			logger.Sugar.Infow("Adapter stopped successfully", "adapter_name", adapter.AdapterName())
		}
	}
	runningAdapters = nil
}

func extractBaseConfig(adapterConfig interface{}) *adapters.ChatAdapterConfig {
	v := reflect.ValueOf(adapterConfig)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	// Iterate through fields to find embedded ChatAdapterConfig
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type() == reflect.TypeOf(adapters.ChatAdapterConfig{}) {
			return field.Addr().Interface().(*adapters.ChatAdapterConfig)
		}
	}

	return nil
}
