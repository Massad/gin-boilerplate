package controllers

import (
	"errors"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/google/uuid"

	"net/http"

	"github.com/gin-gonic/gin"
)

// ArticleController ...
type ArticleController struct{}

var articleModel = new(models.ArticleModel)

// Create Article godoc
// @Summary Create Article example
// @Schemes
// @Description Create Article example
// @Tags Article
// @Accept json
// @Produce json
// @Param article body forms.CreateArticleForm true "Article"
// @Success 	 200  {object}  forms.ArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article [post]
func (ctrl ArticleController) Create(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	var form forms.CreateArticleForm

	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := validate.Translate(validationErr, forms.ArticleMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	id, err := articleModel.Create(c.Request.Context(), tenantID, userID, form)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Article could not be created"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article created", "id": id.String()})
}

// Get All Articles godoc
// @Summary Get All Articles example
// @Schemes
// @Description Get All Articles example
// @Tags Article
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.AllArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /articles [GET]
func (ctrl ArticleController) All(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	results, err := articleModel.All(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"Message": "Could not get articles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// Get One Article godoc
// @Summary Get One Article example
// @Schemes
// @Description One All Article example
// @Tags Article
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.OneArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article/{id} [GET]
func (ctrl ArticleController) One(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	id := c.Param("id")

	getID, err := uuid.Parse(id)
	if err != nil || getID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	data, err := articleModel.One(c.Request.Context(), tenantID, userID, getID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Update Article godoc
// @Summary Update Article example
// @Schemes
// @Description Update Article example
// @Tags Article
// @Accept json
// @Produce json
// @Param article body forms.CreateArticleForm true "Article"
// @Success 	 200  {object}  models.ArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article/{id} [PUT]
func (ctrl ArticleController) Update(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	id := c.Param("id")

	getID, err := uuid.Parse(id)
	if err != nil || getID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	var form forms.CreateArticleForm

	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := validate.Translate(validationErr, forms.ArticleMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	err = articleModel.Update(c.Request.Context(), tenantID, userID, getID, form)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Article not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"Message": "Article could not be updated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article updated"})
}

// Delete Article godoc
// @Summary Delete Article example
// @Schemes
// @Description Delete Article example
// @Tags Article
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.OneArticleResponse
// @Success 	 404  {object}  forms.ArticleResponse
// @Failure      406  {object}  forms.ArticleResponse
// @Router /article/{id} [DELETE]
func (ctrl ArticleController) Delete(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	id := c.Param("id")

	getID, err := uuid.Parse(id)
	if err != nil || getID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	err = articleModel.Delete(c.Request.Context(), tenantID, userID, getID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Article not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"Message": "Article could not be deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}
