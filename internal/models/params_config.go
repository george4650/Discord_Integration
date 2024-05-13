package models

type Params struct {
	DiscordServerURL        string `json:"discord_server_url"`          // url ссылка для редиректа пользователя в случае прохождения oauth2 аут-ции
	RedirectURLIfRejectAuth string `json:"redirect_url_if_reject_auth"` // url ссылка для редиректа пользователя в случае отмены oauth2 аут-ции
}
