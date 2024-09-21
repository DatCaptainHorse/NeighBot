package discord

import (
	"NeighBot/adapters"
	"NeighBot/logger"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type DiscordConfig struct {
	adapters.ChatAdapterConfig
	Token string `json:"token"`
}

type DiscordAdapter struct {
	config  DiscordConfig
	session *discordgo.Session
}

var responding = false
var usersCache = make(map[string]*discordgo.User)

func (d *DiscordAdapter) SetConfig(cfg interface{}) error {
	c, ok := cfg.(*DiscordConfig)
	if !ok {
		return errors.New("invalid config type for DiscordAdapter")
	}
	d.config = *c
	return nil
}

func (d *DiscordAdapter) Initialize() error {
	if d.config.Token == "" {
		return errors.New("discord token is required")
	}

	session, err := discordgo.New("Bot " + d.config.Token)
	if err != nil {
		return err
	}

	d.session = session
	d.session.AddHandler(d.messageCreateHandler)
	d.session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	logger.Sugar.Infow("Discord session initialized", "adapter", d.AdapterName())
	return nil
}

func (d *DiscordAdapter) Start() error {
	if d.session == nil {
		return errors.New("discord session not initialized")
	}

	if err := d.session.Open(); err != nil {
		return err
	}

	logger.Sugar.Infow("Discord connection established", "adapter", d.AdapterName())
	return nil
}

func (d *DiscordAdapter) Stop() error {
	if d.session == nil {
		return errors.New("discord session not initialized")
	}

	if err := d.session.Close(); err != nil {
		return err
	}

	logger.Sugar.Infow("Discord connection closed", "adapter", d.AdapterName())
	return nil
}

func (d *DiscordAdapter) AdapterName() string {
	return "discord"
}

func (d *DiscordAdapter) messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Fetch context by channel ID (TODO: combine server + channel ID to be sure?)
	ctx := d.config.MemoryStore.GetContextForChat(m.ChannelID)
	if ctx == nil {
		// Skip unknown chats
		return
	}

	// Cache the user
	usersCache[m.Author.GlobalName] = m.Author

	// Get server and channel names to construct source
	server, err := s.State.Guild(m.GuildID)
	if err != nil {
		// Try with GET
		server, err = s.Guild(m.GuildID)
		if err != nil {
			logger.Sugar.Errorw("Failed to get guild", "error", err)
			return
		}
	}
	serverName := server.Name

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Try with GET
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			logger.Sugar.Errorw("Failed to get channel", "error", err)
			return
		}
	}
	channelName := channel.Name

	source := fmt.Sprintf("%s:%s:%s", d.AdapterName(), serverName, channelName)

	// Log the incoming message
	logger.Sugar.Infow("Incoming message",
		"author", m.Author.Username,
		"content", m.Content,
		"channel_id", m.ChannelID,
	)

	// Replace the bot mention in message content with '@NeighBot' for proper formatting
	formatted := strings.ReplaceAll(m.Content, s.State.User.Mention(), "@NeighBot")

	// Add user message
	if err = d.config.MemoryStore.AddUserMessage(ctx.ID, source, m.Author.GlobalName, formatted); err != nil {
		logger.Sugar.Errorw("Failed to add user message", "error", err)
		return
	}

	// Check if the bot is mentioned
	if !strings.Contains(m.Content, s.State.User.Mention()) {
		return
	}

	// If already responding, skip
	if responding {
		return
	}

	// Set typing state
	if err = s.ChannelTyping(m.ChannelID); err != nil {
		// Just warn, no need to stop the process
		logger.Sugar.Warnw("Failed to set typing state", "error", err)
	}

	responding = true

	// Generate response from LLM
	messages := ctx.Messages
	response, err := d.config.LLMClient.GenerateResponse(context.Background(), messages)
	if err != nil {
		logger.Sugar.Errorw("Failed to generate response", "error", err)
		responding = false
		return
	}

	responding = false

	// Apply filters to the response
	response = ctx.ApplyFilters(response)

	// Log the response
	logger.Sugar.Infow("Generated response",
		"content", response,
		"channel_id", m.ChannelID,
	)

	// Add response
	if err = d.config.MemoryStore.AddAssistantMessage(ctx.ID, source, response); err != nil {
		logger.Sugar.Errorw("Failed to add assistant message", "error", err)
		return
	}

	// Replace any mentions in response '@user name' with Discord format <@!user ID>
	if strings.Contains(response, "@") {
		for k, v := range usersCache {
			response = strings.ReplaceAll(response, "@"+k, v.Mention())
		}
	}

	// Send the response to Discord, if over 2000 characters, send in chunks
	if len(response) > 2000 {
		for i := 0; i < len(response); i += 2000 {
			end := i + 2000
			if end > len(response) {
				end = len(response)
			}
			_, err = s.ChannelMessageSend(m.ChannelID, response[i:end])
			if err != nil {
				logger.Sugar.Errorw("Failed to send response", "error", err)
			}
		}
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, response)
		if err != nil {
			logger.Sugar.Errorw("Failed to send response", "error", err)
		}
	}
}
