package forms

//ArticleForm ...
type ArticleForm struct {
	Title   string `form:"title" json:"title" binding:"required"`
	Content string `form:"content" json:"content" binding:"required"`
}
