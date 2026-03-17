package forms

// CreateArticleForm ...
type CreateArticleForm struct {
	Title   string `form:"title" json:"title" binding:"required,min=3,max=100"`
	Content string `form:"content" json:"content" binding:"required,min=3,max=1000"`
}

// ArticleResponse ...
type ArticleResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// ArticleMessages defines validation error messages for article forms.
var ArticleMessages = ValidationMessages{
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
