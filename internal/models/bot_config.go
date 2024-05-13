package models

type BotConfig struct {
	DiscordBotToken          string `json:"discord_bot_token"`
	DiscordBotClientID       string `json:"discord_bot_client_id"`
	DiscordBotClientSecret   string `json:"discord_bot_client_secret"`
	DiscordOAUTH2ApiCallback string `json:"discord_oauth2_api_callback"` // ручка на которую будут направлены данные об участнике прошедшего oauth2 аут-цию
	DiscordServerID          int    `json:"discord_server_id"`
}
