package forms

import (
	validate "github.com/Massad/gin-boilerplate/internal/platform/validate"
)

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
