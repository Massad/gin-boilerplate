package admin

import (
	validate "github.com/Massad/gin-boilerplate/internal/platform/validate"
)

// LoginForm ...
type LoginForm struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=50"`
}

// LoginMessages defines validation error messages for the login form.
var LoginMessages = validate.ValidationMessages{
	"Email": {
		"required": "Please enter your email",
		"email":    "Please enter a valid email",
	},
	"Password": {
		"required": "Please enter your password",
		"min":      "Your password should be between 3 and 50 characters",
		"max":      "Your password should be between 3 and 50 characters",
		"eqfield":  "Your passwords does not match",
	},
}

// TenantStatusForm carries the platform tenant status toggle: 1 active,
// 0 suspended.
type TenantStatusForm struct {
	Status *int16 `form:"status" json:"status" binding:"required,oneof=0 1"`
}

// AdminStatusMessages defines validation error messages for the status form.
var AdminStatusMessages = validate.ValidationMessages{
	"Status": {
		"required": "Please provide a status",
		"oneof":    "Status must be 0 (suspended) or 1 (active)",
	},
}
