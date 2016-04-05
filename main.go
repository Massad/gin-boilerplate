package main

import (
	"fmt"
	"gin-boilerplate/controllers"
	"gin-boilerplate/db"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

//CORSMiddleware ...
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

func main() {
	r := gin.Default()

	store, _ := sessions.NewRedisStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	r.Use(sessions.Sessions("gin-boilerplate-session", store))

	r.Use(CORSMiddleware())

	db.Init()

	v1 := r.Group("/v1")
	{
		/*** START USER ***/
		v1.POST("/user/login", controllers.Login)
		v1.POST("/user/register", controllers.Register)
		v1.GET("/user/logout", controllers.Logout)

		/*** START Article ***/
		v1.POST("/article", controllers.CreateArticle)
		v1.GET("/article/:id", controllers.GetArticle)
		v1.GET("/articles", controllers.GetArticles)
		v1.PUT("/article/:id", controllers.UpdateArticle)
		v1.DELETE("/article/:id", controllers.DeleteArticle)
	}

	r.Static("/public", "./public")

	r.Run(":9000")
}
