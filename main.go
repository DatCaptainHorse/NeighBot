package main

import (
	"NeighBot/config"
	"NeighBot/filters"
	"NeighBot/llm"
	"NeighBot/logger"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// Initialize logger
	if err := logger.InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.SyncLogger()

	// Define flags
	configDirFlag := flag.String("config-dir", "", "Path to configuration directory")
	flag.Parse()

	// Determine config directory by priority: ENV > Flag > Default
	var configDir string
	if configDirEnv := os.Getenv("CONFIG_DIR"); configDirEnv != "" {
		configDir = configDirEnv
	} else if *configDirFlag != "" {
		configDir = *configDirFlag
	} else {
		panic("No config directory specified")
	}

	logger.Sugar.Infow("Starting NeighBot",
		"config_dir", configDir,
	)
	// Ensure the directories exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Sugar.Fatalw("Failed to create config directory",
			"config_dir", configDir,
			"error", err,
		)
	}

	mainConfigFile := filepath.Join(configDir, "main.json")

	var mainConfig config.MainConfig
	if _, err := os.Stat(mainConfigFile); os.IsNotExist(err) {
		if err = mainConfig.CreateDefault(mainConfigFile); err != nil {
			logger.Sugar.Fatalw("Failed to create default config", "error", err)
		}
		logger.Sugar.Infow("Created default config", "file", mainConfigFile)
	} else {
		if err = mainConfig.Load(mainConfigFile); err != nil {
			logger.Sugar.Fatalw("Failed to load config", "error", err)
		}
	}

	// Ensure the data directory exists
	dataDir := filepath.Join(configDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Sugar.Fatalw("Failed to create data directory", "data_dir", dataDir, "error", err)
	}

	// Initialize the memory store
	memoryStore := llm.NewMemoryStore(dataDir)
	if err := memoryStore.LoadAllContexts(); err != nil {
		logger.Sugar.Fatalw("Failed to load contexts", "error", err)
	}

	if err := memoryStore.PopulateEmptyFolders(); err != nil {
		logger.Sugar.Fatalw("Failed to populate empty folders", "error", err)
	}

	// Initialize the LLM client
	llmClient := llm.NewOpenAIClient(
		mainConfig.LLM.APIKey,
		mainConfig.LLM.Endpoint,
		mainConfig.LLM.Model,
	)
	logger.Sugar.Infow("LLM client initialized",
		"endpoint", mainConfig.LLM.Endpoint,
		"model", mainConfig.LLM.Model,
	)

	// Initialize filters
	filters.InitializeFilters()

	// Initialize adapters
	if err := HandleAdapters(&mainConfig, memoryStore, llmClient); err != nil {
		logger.Sugar.Fatalw("Failed to initialize adapters", "error", err)
	}

	// Create signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Wait for a signal to quit
	<-sc

	// Stopping //
	logger.Sugar.Info("Stopping NeighBot")

	// Stop adapters
	StopAdapters()

	// Saving
	logger.Sugar.Info("Saving contexts")
	if err := memoryStore.SaveAllContexts(); err != nil {
		logger.Sugar.Errorw("Failed to save contexts", "error", err)
	}

	logger.Sugar.Info("Saving config")
	if err := mainConfig.Save(mainConfigFile); err != nil {
		logger.Sugar.Errorw("Failed to save config", "error", err)
	}

	logger.Sugar.Info("NeighBot stopped gracefully")
}
