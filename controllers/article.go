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

	err := models.CreateArticle(userID, articleForm)

	if err != nil {
		c.JSON(406, gin.H{"message": err.Error()})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{"status": 1})
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
		data := models.GetArticle(userID, id)
		if data.ID == 0 {
			c.JSON(404, gin.H{"Message": "Article not found"})
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

	data := models.GetArticles(userID)
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
			c.JSON(406, gin.H{"Message": "Article could not be updated", "err": err.Error()})
			c.Abort()
			return
		}
		c.JSON(200, gin.H{"status": "1"})
	} else {
		c.JSON(404, gin.H{"Message": "Invalid parameter"})
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
			c.JSON(406, gin.H{"Message": "Article could not be deleted", "err": err.Error()})
			c.Abort()
			return
		}
		c.JSON(200, gin.H{"status": "1"})
	} else {
		c.JSON(404, gin.H{"Message": "Invalid parameter"})
	}
}
