package models

type User struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"` // Имя пользователя, установленное вне сервера
	Bot        bool   `json:"bot"`         // Является ли пользователь ботом.
}
