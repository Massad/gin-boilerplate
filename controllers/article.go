package controllers

import (
	"gin-boilerplate/forms"
	"gin-boilerplate/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

//CreateArticle ...
func CreateArticle(c *gin.Context) {
	userID := getUserID(c)

	if userID <= 0 {
		c.JSON(403, gin.H{"message": "Please login first"})
		c.Abort()
		return
	}

	var articleForm forms.ArticleForm

	if c.BindJSON(&articleForm) != nil {
		c.JSON(406, gin.H{"message": "Invalid parameters", "form": articleForm})
		c.Abort()
		return
	}

	articleID, err := models.CreateArticle(userID, articleForm)

	if articleID > 0 && err != nil {
		c.JSON(406, gin.H{"message": "Article could not be created", "error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{"message": "Article created", "id": articleID})
}

//GetArticle ...
func GetArticle(c *gin.Context) {
	userID := getUserID(c)

	if userID <= 0 {
		c.JSON(403, gin.H{"message": "Please login first"})
		c.Abort()
		return
	}

	id := c.Param("id")

	if id, err := strconv.ParseInt(id, 10, 64); err == nil {
		data, err := models.GetArticle(userID, id)
		if err != nil {
			c.JSON(404, gin.H{"Message": "Article not found", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(200, gin.H{"data": data})
	} else {
		c.JSON(404, gin.H{"Message": "Invalid parameter"})
	}
}

//GetArticles ...
func GetArticles(c *gin.Context) {
	userID := getUserID(c)

	if userID <= 0 {
		c.JSON(403, gin.H{"message": "Please login first"})
		c.Abort()
		return
	}

	data, err := models.GetArticles(userID)

	if err != nil {
		c.JSON(406, gin.H{"Message": "Could not get the articles", "error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{"data": data})
}

//UpdateArticle ...
func UpdateArticle(c *gin.Context) {
	userID := getUserID(c)

	if userID <= 0 {
		c.JSON(403, gin.H{"message": "Please login first"})
		c.Abort()
		return
	}

	id := c.Param("id")
	if id, err := strconv.ParseInt(id, 10, 64); err == nil {

		var articleForm forms.ArticleForm

		if c.BindJSON(&articleForm) != nil {
			c.JSON(406, gin.H{"message": "Invalid parameters", "form": articleForm})
			c.Abort()
			return
		}

		err := models.UpdateArticle(userID, id, articleForm)
		if err != nil {
			c.JSON(406, gin.H{"Message": "Article could not be updated", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(200, gin.H{"message": "Article updated"})
	} else {
		c.JSON(404, gin.H{"Message": "Invalid parameter", "error": err.Error()})
	}
}

//DeleteArticle ...
func DeleteArticle(c *gin.Context) {
	userID := getUserID(c)

	if userID <= 0 {
		c.JSON(403, gin.H{"message": "Please login first"})
		c.Abort()
		return
	}

	id := c.Param("id")
	if id, err := strconv.ParseInt(id, 10, 64); err == nil {
		err := models.DeleteArticle(userID, id)
		if err != nil {
			c.JSON(406, gin.H{"Message": "Article could not be deleted", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(200, gin.H{"message": "Article deleted"})
	} else {
		c.JSON(404, gin.H{"Message": "Invalid parameter", "error": err.Error()})
	}
}
