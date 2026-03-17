package forms

// LoginForm ...
type LoginForm struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=50"`
}

// RegisterForm ...
type RegisterForm struct {
	Name     string `form:"name" json:"name" binding:"required,min=3,max=20,fullName"`
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=50"`
}

// LoginMessages defines validation error messages for login form.
var LoginMessages = ValidationMessages{
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

// RegisterMessages defines validation error messages for register form.
var RegisterMessages = ValidationMessages{
	"Name": {
		"required": "Please enter your name",
		"min":      "Your name should be between 3 to 20 characters",
		"max":      "Your name should be between 3 to 20 characters",
		"fullName": "Name should not include any special characters or numbers",
	},
	"Email": {
		"required": "Please enter your email",
		"email":    "Please enter a valid email",
	},
	"Password": {
		"required": "Please enter your password",
		"min":      "Your password should be between 3 and 50 characters",
		"max":      "Your password should be between 3 and 50 characters",
	},
}
