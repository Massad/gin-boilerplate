package controllers

import (
	"strconv"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"

	"net/http"

	"github.com/gin-gonic/gin"
)

//ArticleController ...
type ArticleController struct{}

var articleModel = new(models.ArticleModel)

//Create ...
func (ctrl ArticleController) Create(c *gin.Context) {
	if userID := getUserID(c); userID != 0 {
		var articleForm forms.ArticleForm

		if c.ShouldBindJSON(&articleForm) != nil {
			c.JSON(http.StatusNotAcceptable, gin.H{"message": "Invalid form"})
			c.Abort()
			return
		}

		articleID, err := articleModel.Create(userID, articleForm)

		if articleID == 0 && err != nil {
			c.JSON(http.StatusNotAcceptable, gin.H{"message": "Article could not be created", "error": err.Error()})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article created", "id": articleID})
	}
}

//All ...
func (ctrl ArticleController) All(c *gin.Context) {
	if userID := getUserID(c); userID != 0 {

		data, err := articleModel.All(userID)

		if err != nil {
			c.JSON(http.StatusNotAcceptable, gin.H{"Message": "Could not get articles", "error": err.Error()})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": data})
	}
}

//One ...
func (ctrl ArticleController) One(c *gin.Context) {
	if userID := getUserID(c); userID != 0 {

		id := c.Param("id")
		if id, err := strconv.ParseInt(id, 10, 64); err == nil {

			data, err := articleModel.One(userID, id)

			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"Message": "Article not found", "error": err.Error()})
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": data})

		} else {
			c.JSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		}
	}
}

//Update ...
func (ctrl ArticleController) Update(c *gin.Context) {
	if userID := getUserID(c); userID != 0 {

		id := c.Param("id")
		if id, err := strconv.ParseInt(id, 10, 64); err == nil {

			var articleForm forms.ArticleForm

			if c.ShouldBindJSON(&articleForm) != nil {
				c.JSON(http.StatusNotAcceptable, gin.H{"message": "Invalid form"})
				c.Abort()
				return
			}

			err := articleModel.Update(userID, id, articleForm)
			if err != nil {
				c.JSON(http.StatusNotAcceptable, gin.H{"Message": "Article could not be updated", "error": err.Error()})
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Article updated"})

		} else {
			c.JSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter", "error": err.Error()})
		}
	}
}

//Delete ...
func (ctrl ArticleController) Delete(c *gin.Context) {
	if userID := getUserID(c); userID != 0 {

		id := c.Param("id")
		if id, err := strconv.ParseInt(id, 10, 64); err == nil {

			err := articleModel.Delete(userID, id)
			if err != nil {
				c.JSON(http.StatusNotAcceptable, gin.H{"Message": "Article could not be deleted", "error": err.Error()})
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})

		} else {
			c.JSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		}
	}
}
