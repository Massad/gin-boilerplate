package article

// AllArticleResponse represents the top-level list structure.
type AllArticleResponse struct {
	Results []Result `json:"results"`
}

// OneArticleResponse wraps a single article.
type OneArticleResponse struct {
	Data ArticleResponse `json:"data"`
}

// Result represents each result item
type Result struct {
	Data []ArticleResponse `json:"data"`
	Meta Meta              `json:"meta"`
}

// ArticleResponse represents an individual article
type ArticleResponse struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	UpdatedAt int64        `json:"updated_at"`
	CreatedAt int64        `json:"created_at"`
	User      UserResponse `json:"user"`
}

// UserResponse represents the article's author
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Meta represents metadata for pagination or total count
type Meta struct {
	Total int `json:"total"`
}
