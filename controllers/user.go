package controllers

import (
	"gin-boilerplate/forms"
	"gin-boilerplate/models"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

//Login ...
func Login(c *gin.Context) {
	var loginForm forms.LoginForm

	if c.BindJSON(&loginForm) == nil {

		user, err := models.Login(loginForm)
		if err == nil {
			session := sessions.Default(c)
			session.Set("user_id", user.ID)
			session.Save()
			c.JSON(200, gin.H{"message": "User logged in", "user": user})
		} else {
			c.JSON(406, gin.H{"message": "Invalid login details", "error": err.Error()})
		}
	} else {
		c.JSON(406, gin.H{"message": "Invalid login details", "form": loginForm})
	}
}

//Register ...
func Register(c *gin.Context) {
	var registerForm forms.RegisterForm

	if c.BindJSON(&registerForm) == nil {
		user, err := models.Register(registerForm)

		if err != nil {
			c.JSON(406, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		if user.ID > 0 {
			session := sessions.Default(c)
			session.Set("userID", user.ID)
			session.Save()
			c.JSON(200, gin.H{"message": "Success register", "user": user})
		} else {
			c.JSON(406, gin.H{"message": "Could not create a user", "error": err.Error()})
		}
	} else {
		c.JSON(406, gin.H{"message": "Invalid register details", "form": registerForm})
	}
}

//Logout ...
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Set("user_id", nil)
	session.Save()
	c.JSON(200, nil)
}

//getUserID ...
func getUserID(c *gin.Context) int64 {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID != nil {
		var i64 int64
		i64 = int64(userID.(int))
		return i64
	}
	return 0
}

//GetUsers ...
func GetUsers(c *gin.Context) {
	data, _ := models.GetUsers()
	c.JSON(200, data)
}
