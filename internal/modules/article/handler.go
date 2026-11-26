package article

import (
	"errors"
	"net/http"

	"github.com/Massad/gin-boilerplate/internal/platform/httpx"
	"github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves the /v1 article HTTP surface.
type Handler struct {
	articles *Service
}

// NewHandler builds the article handler on its service.
func NewHandler(articles *Service) *Handler {
	return &Handler{articles: articles}
}

// getUserID reads the authenticated user id set by TokenAuth.
func getUserID(c *gin.Context) (userID uuid.UUID) {
	userID, _ = uuid.Parse(c.GetString("userID"))
	return userID
}

// getTenantID reads the authenticated tenant id set by TokenAuth.
func getTenantID(c *gin.Context) (tenantID uuid.UUID) {
	tenantID, _ = uuid.Parse(c.GetString("tenantID"))
	return tenantID
}

// Create Article godoc
// @Summary Create Article example
// @Schemes
// @Description Create Article example
// @Tags Article
// @Accept json
// @Produce json
// @Param article body article.CreateArticleForm true "Article"
// @Success 	 200  {object}  article.ArticleResponse
// @Failure      406  {object}  article.ArticleResponse
// @Router /article [post]
func (h Handler) Create(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	var form CreateArticleForm

	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := validate.Translate(validationErr, ArticleMessages)
		httpx.AbortMessage(c, http.StatusNotAcceptable, message)
		return
	}

	id, err := h.articles.Create(c.Request.Context(), tenantID, userID, form)
	if err != nil {
		httpx.AbortMessage(c, http.StatusNotAcceptable, "Article could not be created")
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
// @Success 	 200  {object}  article.AllArticleResponse
// @Failure      406  {object}  article.ArticleResponse
// @Router /articles [GET]
func (h Handler) All(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	results, err := h.articles.All(c.Request.Context(), tenantID, userID)
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
// @Success 	 200  {object}  article.OneArticleResponse
// @Failure      406  {object}  article.ArticleResponse
// @Router /article/{id} [GET]
func (h Handler) One(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	getID, err := httpx.ParseUUIDParam(c, "id")
	if err != nil || getID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	data, err := h.articles.One(c.Request.Context(), tenantID, userID, getID)
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
// @Param article body article.CreateArticleForm true "Article"
// @Success 	 200  {object}  article.ArticleResponse
// @Failure      406  {object}  article.ArticleResponse
// @Router /article/{id} [PUT]
func (h Handler) Update(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	getID, err := httpx.ParseUUIDParam(c, "id")
	if err != nil || getID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	var form CreateArticleForm

	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := validate.Translate(validationErr, ArticleMessages)
		httpx.AbortMessage(c, http.StatusNotAcceptable, message)
		return
	}

	err = h.articles.Update(c.Request.Context(), tenantID, userID, getID, form)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Article not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"Message": "Article could not be updated"})
		return
	}

	httpx.JSONMessage(c, http.StatusOK, "Article updated")
}

// Delete Article godoc
// @Summary Delete Article example
// @Schemes
// @Description Delete Article example
// @Tags Article
// @Accept json
// @Produce json
// @Success 	 200  {object}  article.OneArticleResponse
// @Success 	 404  {object}  article.ArticleResponse
// @Failure      406  {object}  article.ArticleResponse
// @Router /article/{id} [DELETE]
func (h Handler) Delete(c *gin.Context) {
	userID := getUserID(c)
	tenantID := getTenantID(c)

	getID, err := httpx.ParseUUIDParam(c, "id")
	if err != nil || getID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	err = h.articles.Delete(c.Request.Context(), tenantID, userID, getID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Article not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"Message": "Article could not be deleted"})
		return
	}

	httpx.JSONMessage(c, http.StatusOK, "Article deleted")
}
