// Package article serves the tenant article CRUD resource: the /v1 article
// HTTP surface, its use cases and its tenant-backed repository.
package article

import (
	validate "github.com/Massad/gin-boilerplate/internal/platform/validate"
)

// CreateArticleForm is the create/update payload, validated with binding tags.
type CreateArticleForm struct {
	Title   string `form:"title" json:"title" binding:"required,min=3,max=100"`
	Content string `form:"content" json:"content" binding:"required,min=3,max=1000"`
}

// ArticleMessages defines validation error messages for article forms.
var ArticleMessages = validate.ValidationMessages{
	"Title": {
		"required": "Please enter the article title",
		"min":      "Title should be between 3 to 100 characters",
		"max":      "Title should be between 3 to 100 characters",
	},
	"Content": {
		"required": "Please enter the article content",
		"min":      "Content should be between 3 to 1000 characters",
		"max":      "Content should be between 3 to 1000 characters",
	},
}
