package models

type BanReason struct {
	Reason *string `json:"reason" binding:"omitempty"`
}
