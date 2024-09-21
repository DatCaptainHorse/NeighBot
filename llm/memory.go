package llm

import (
	"NeighBot/logger"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type MemoryStore struct {
	contexts map[string]*StoredContext
	dataDir  string
}

func NewMemoryStore(dataDir string) *MemoryStore {
	return &MemoryStore{
		contexts: make(map[string]*StoredContext),
		dataDir:  dataDir,
	}
}

func (m *MemoryStore) CreateContext(contextID string) *StoredContext {
	newContext := &StoredContext{
		ID:          contextID,
		Name:        "New Chat",
		Description: "New context for testing.",
		Filters: map[string]bool{
			"remove_emojis":   true,
			"remove_emphasis": true,
			"remove_links":    true,
		},
		AssociatedChats: []string{},
	}

	m.AddContext(newContext)
	return newContext
}

func (m *MemoryStore) AddContext(ctx *StoredContext) {
	if _, exists := m.contexts[ctx.ID]; exists {
		logger.Sugar.Warnw("Context already exists, skipping addition", "context_id", ctx.ID)
		return
	}

	m.contexts[ctx.ID] = ctx
	logger.Sugar.Infow("Context added to MemoryStore", "context_id", ctx.ID)

	if err := m.SaveContextConfig(ctx); err != nil {
		logger.Sugar.Errorw("Failed to save context config during AddContext", "context_id", ctx.ID, "error", err)
	}
	if err := m.SaveContextMemory(ctx); err != nil {
		logger.Sugar.Errorw("Failed to save context memory during AddContext", "context_id", ctx.ID, "error", err)
	}
}

func (m *MemoryStore) GetAllContextIDs() []string {
	contextIDs := make([]string, 0, len(m.contexts))
	for contextID := range m.contexts {
		contextIDs = append(contextIDs, contextID)
	}
	return contextIDs
}

func (m *MemoryStore) GetContext(contextID string) *StoredContext {
	if ctx, exists := m.contexts[contextID]; exists {
		return ctx
	}
	return nil
}

func (m *MemoryStore) GetContextForChat(chatID string) *StoredContext {
	for _, ctx := range m.contexts {
		for _, chat := range ctx.AssociatedChats {
			if chat == chatID {
				return ctx
			}
		}
	}
	return nil
}

func (m *MemoryStore) SaveContextConfig(ctx *StoredContext) error {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		logger.Sugar.Errorw("Failed to marshal context config for saving", "context_id", ctx.ID, "error", err)
		return err
	}

	path := filepath.Join(m.dataDir, ctx.ID, "config.json")
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logger.Sugar.Errorw("Failed to create directory for context config", "context_id", ctx.ID, "error", err)
		return err
	}

	if err = os.WriteFile(path, data, 0644); err != nil {
		logger.Sugar.Errorw("Failed to save context config to file", "context_id", ctx.ID, "error", err)
		return err
	}

	logger.Sugar.Infow("Successfully saved context config", "context_id", ctx.ID)
	return nil
}

func (m *MemoryStore) SaveContextMemory(ctx *StoredContext) error {
	memoryData := map[string]interface{}{
		"messages": ctx.Messages,
	}

	data, err := json.MarshalIndent(memoryData, "", "  ")
	if err != nil {
		logger.Sugar.Errorw("Failed to marshal context memory for saving", "context_id", ctx.ID, "error", err)
		return err
	}

	path := filepath.Join(m.dataDir, ctx.ID, "memory.json")
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logger.Sugar.Errorw("Failed to create directory for context memory", "context_id", ctx.ID, "error", err)
		return err
	}

	if err = os.WriteFile(path, data, 0644); err != nil {
		logger.Sugar.Errorw("Failed to save context memory to file", "context_id", ctx.ID, "error", err)
		return err
	}

	logger.Sugar.Infow("Successfully saved context memory", "context_id", ctx.ID)
	return nil
}

func (m *MemoryStore) LoadContextConfig(contextID string) (*StoredContext, error) {
	path := filepath.Join(m.dataDir, contextID, "config.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Sugar.Warnw("Context config file does not exist", "context_id", contextID)
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Sugar.Errorw("Failed to read context config file", "context_id", contextID, "error", err)
		return nil, err
	}

	context := &StoredContext{ID: contextID}
	if err = json.Unmarshal(data, &context); err != nil {
		logger.Sugar.Errorw("Failed to unmarshal context config file", "context_id", contextID, "error", err)
		return nil, err
	}

	return context, nil
}

func (m *MemoryStore) LoadContextMemory(contextID string, ctx *StoredContext) error {
	path := filepath.Join(m.dataDir, contextID, "memory.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Sugar.Warnw("Context memory file does not exist", "context_id", contextID)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Sugar.Errorw("Failed to read context memory file", "context_id", contextID, "error", err)
		return err
	}

	var memoryData struct {
		Messages []StoredMessage `json:"messages"`
	}
	if err = json.Unmarshal(data, &memoryData); err != nil {
		logger.Sugar.Errorw("Failed to unmarshal context memory file", "context_id", contextID, "error", err)
		return err
	}

	ctx.Messages = memoryData.Messages
	logger.Sugar.Infow("Successfully loaded context memory", "context_id", contextID)
	return nil
}

func (m *MemoryStore) LoadAllContexts() error {
	dir := filepath.Join(m.dataDir)
	files, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		logger.Sugar.Errorw("Failed to read contexts directory", "path", dir, "error", err)
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			contextID := f.Name()
			ctx, err := m.LoadContextConfig(contextID)
			if err != nil {
				return err
			}
			if ctx == nil {
				continue
			}

			if err = m.LoadContextMemory(contextID, ctx); err != nil {
				return err
			}

			m.contexts[contextID] = ctx
		}
	}

	logger.Sugar.Infow("Successfully loaded all contexts")
	return nil
}

func (m *MemoryStore) SaveAllContexts() error {
	for contextID, ctx := range m.contexts {
		if err := m.SaveContextConfig(ctx); err != nil {
			logger.Sugar.Errorw("Failed to save context config", "context_id", contextID, "error", err)
			return err
		}
		if err := m.SaveContextMemory(ctx); err != nil {
			logger.Sugar.Errorw("Failed to save context memory", "context_id", contextID, "error", err)
			return err
		}
	}
	logger.Sugar.Infow("Successfully saved all contexts")
	return nil
}

func (m *MemoryStore) PopulateEmptyFolders() error {
	dir := filepath.Join(m.dataDir)
	files, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		logger.Sugar.Errorw("Failed to read contexts directory", "path", dir, "error", err)
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			contextID := f.Name()
			configPath := filepath.Join(dir, contextID, "config.json")
			memoryPath := filepath.Join(dir, contextID, "memory.json")

			if _, err = os.Stat(configPath); os.IsNotExist(err) {
				configData := map[string]interface{}{
					"context_id":  contextID,
					"name":        "New Chat",
					"description": "New context for testing.",
					"filters": map[string]bool{
						"remove_emojis":   true,
						"remove_emphasis": true,
						"remove_links":    true,
					},
					"associated_chats": []string{},
				}

				data, err := json.MarshalIndent(configData, "", "  ")
				if err != nil {
					logger.Sugar.Errorw("Failed to marshal context config for saving", "context_id", contextID, "error", err)
					return err
				}

				if err = os.WriteFile(configPath, data, 0644); err != nil {
					logger.Sugar.Errorw("Failed to save context config to file", "context_id", contextID, "error", err)
					return err
				}
			}

			if _, err = os.Stat(memoryPath); os.IsNotExist(err) {
				memoryData := map[string]interface{}{
					"messages": []StoredMessage{},
				}

				data, err := json.MarshalIndent(memoryData, "", "  ")
				if err != nil {
					logger.Sugar.Errorw("Failed to marshal context memory for saving", "context_id", contextID, "error", err)
					return err
				}

				if err = os.WriteFile(memoryPath, data, 0644); err != nil {
					logger.Sugar.Errorw("Failed to save context memory to file", "context_id", contextID, "error", err)
					return err
				}
			}
		}
	}

	logger.Sugar.Infow("Successfully populated empty folders")
	return nil
}

func (m *MemoryStore) AddUserMessage(contextID, source, username, content string) error {
	ctx := m.GetContext(contextID)
	if ctx == nil {
		logger.Sugar.Warnw("Context does not exist", "context_id", contextID)
		return nil
	}

	userMessage := StoredMessage{
		Username:  username,
		Source:    source,
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}
	ctx.AddMessage(userMessage)
	return m.SaveContextMemory(ctx)
}

func (m *MemoryStore) AddAssistantMessage(contextID, source, content string) error {
	ctx := m.GetContext(contextID)
	if ctx == nil {
		logger.Sugar.Warnw("Context does not exist", "context_id", contextID)
		return nil
	}

	assistantMessage := StoredMessage{
		Username:  "NeighBot",
		Source:    source,
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}
	ctx.AddMessage(assistantMessage)
	return m.SaveContextMemory(ctx)
}

func (m *MemoryStore) ApplyFilters(contextID, content string) string {
	ctx := m.GetContext(contextID)
	if ctx == nil {
		logger.Sugar.Warnw("Context does not exist", "context_id", contextID)
		return content
	}

	return ctx.ApplyFilters(content)
}
