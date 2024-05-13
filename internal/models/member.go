package models

import "time"

type Member struct {
	JoinedAt time.Time `json:"joined_at"`
	Nick     string    `json:"nick"` // Фамилия Имя пользователя, установленное на сервере
	User     User      `json:"user"`
	Roles    []string  `json:"roles"`
}

type AddMemberParams struct {
	AccessToken string   `json:"access_token"`
	Nick        string   `json:"nick"`
	Roles       []string `json:"roles"`
}
