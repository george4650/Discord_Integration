package common

const (
	AuthURL  = "https://discord.com/oauth2/authorize"
	TokenURL = "https://discord.com/api/oauth2/token"
)

var (
	Scopes = []string{"identify", "guilds.join"}
)
