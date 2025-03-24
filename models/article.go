package models

import (
	"errors"

	"github.com/Massad/gin-boilerplate/db"
	"github.com/Massad/gin-boilerplate/forms"
)

// Response represents the top-level structure
type AllArticleResponse struct {
	Results []Result `json:"results"`
}

type OneArticleResponse struct {
	Data ArticleResponse `json:"data"`
}

// Result represents each result item
type Result struct {
	Data []ArticleResponse `json:"data"`
	Meta Meta              `json:"meta"`
}

// Article represents an individual article
type ArticleResponse struct {
	ID        int          `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	UpdatedAt int64        `json:"updated_at"`
	CreatedAt int64        `json:"created_at"`
	User      UserResponse `json:"user"`
}

// User represents the article's author
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Meta represents metadata for pagination or total count
type Meta struct {
	Total int `json:"total"`
}

// Article ...
type Article struct {
	ID        int64    `db:"id, primarykey, autoincrement" json:"id"`
	UserID    int64    `db:"user_id" json:"-"`
	Title     string   `db:"title" json:"title"`
	Content   string   `db:"content" json:"content"`
	UpdatedAt int64    `db:"updated_at" json:"updated_at"`
	CreatedAt int64    `db:"created_at" json:"created_at"`
	User      *JSONRaw `db:"user" json:"user"`
}

// ArticleModel ...
type ArticleModel struct{}

// Create ...
func (m ArticleModel) Create(userID int64, form forms.CreateArticleForm) (articleID int64, err error) {
	err = db.GetDB().QueryRow("INSERT INTO public.article(user_id, title, content) VALUES($1, $2, $3) RETURNING id", userID, form.Title, form.Content).Scan(&articleID)
	return articleID, err
}

// One ...
func (m ArticleModel) One(userID, id int64) (article Article, err error) {
	err = db.GetDB().SelectOne(&article, "SELECT a.id, a.title, a.content, a.updated_at, a.created_at, json_build_object('id', u.id, 'name', u.name, 'email', u.email) AS user FROM public.article a LEFT JOIN public.user u ON a.user_id = u.id WHERE a.user_id=$1 AND a.id=$2 LIMIT 1", userID, id)
	return article, err
}

// All ...
func (m ArticleModel) All(userID int64) (articles []DataList, err error) {
	_, err = db.GetDB().Select(&articles, "SELECT COALESCE(array_to_json(array_agg(row_to_json(d))), '[]') AS data, (SELECT row_to_json(n) FROM ( SELECT count(a.id) AS total FROM public.article AS a WHERE a.user_id=$1 LIMIT 1 ) n ) AS meta FROM ( SELECT a.id, a.title, a.content, a.updated_at, a.created_at, json_build_object('id', u.id, 'name', u.name, 'email', u.email) AS user FROM public.article a LEFT JOIN public.user u ON a.user_id = u.id WHERE a.user_id=$1 ORDER BY a.id DESC) d", userID)
	return articles, err
}

// Update ...
func (m ArticleModel) Update(userID int64, id int64, form forms.CreateArticleForm) (err error) {
	//METHOD 1
	//Check the article by ID using this way
	// _, err = m.One(userID, id)
	// if err != nil {
	// 	return err
	// }

	operation, err := db.GetDB().Exec("UPDATE public.article SET title=$2, content=$3 WHERE id=$1", id, form.Title, form.Content)
	if err != nil {
		return err
	}

	success, _ := operation.RowsAffected()
	if success == 0 {
		return errors.New("updated 0 records")
	}

	return err
}

// Delete ...
func (m ArticleModel) Delete(userID, id int64) (err error) {

	operation, err := db.GetDB().Exec("DELETE FROM public.article WHERE id=$1", id)
	if err != nil {
		return err
	}

	success, _ := operation.RowsAffected()
	if success == 0 {
		return errors.New("no records were deleted")
	}

	return err
}
