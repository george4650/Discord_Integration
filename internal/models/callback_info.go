package models

type CallBackInfo struct {
	Code             *string `form:"code" binding:"omitempty,max=50"`
	State            string  `form:"state" binding:"required,max=600"`
	Error            *string `form:"error" binding:"omitempty,max=50"`
	ErrorDescription *string `form:"error_description" binding:"omitempty,max=200"`
}
