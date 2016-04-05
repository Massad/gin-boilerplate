package models

import (
	"errors"
	"fmt"
	"gin-boilerplate/db"
	"gin-boilerplate/forms"
	"time"
)

//Article ...
type Article struct {
	ID        int64    `db:"id, primarykey, autoincrement" json:"id"`
	UserID    int64    `db:"user_id" json:"-"`
	Title     string   `db:"title" json:"title"`
	Content   string   `db:"content" json:"content"`
	UpdatedAt int64    `db:"updated_at" json:"updated_at"`
	CreatedAt int64    `db:"created_at" json:"created_at"`
	User      *JSONRaw `db:"user" json:"user"`
}

//CreateArticle ...
func CreateArticle(userID int64, form forms.ArticleForm) (articleID int64, err error) {
	getDb := db.GetDB()

	checkUser := CheckUser(userID)

	if checkUser != nil {
		return 0, errors.New("User doesn't exist")
	}

	_, err = getDb.Exec("INSERT INTO article(user_id, title, content, updated_at, created_at) VALUES($1, $2, $3, $4, $5) RETURNING id", userID, form.Title, form.Content, time.Now().Unix(), time.Now().Unix())

	if err != nil {
		return 0, err
	}

	articleID, err = getDb.SelectInt("SELECT id FROM article WHERE user_id=$1 ORDER BY id DESC LIMIT 1", userID)

	return articleID, err
}

//GetArticle ...
func GetArticle(userID, id int64) (article Article, err error) {
	err = db.GetDB().SelectOne(&article, "SELECT a.id, a.title, a.content, a.updated_at, a.created_at, json_build_object('id', u.id, 'name', u.name, 'email', u.email) AS user FROM article a LEFT JOIN public.user u ON a.user_id = u.id WHERE a.user_id=$1 AND a.id=$2 GROUP BY a.id, a.title, a.content, a.updated_at, a.created_at, u.id, u.name, u.email LIMIT 1", userID, id)
	return article, err
}

//GetArticles ...
func GetArticles(userID int64) (articles []Article, err error) {
	_, err = db.GetDB().Select(&articles, "SELECT a.id, a.title, a.content, a.updated_at, a.created_at, json_build_object('id', u.id, 'name', u.name, 'email', u.email) AS user FROM article a LEFT JOIN public.user u ON a.user_id = u.id WHERE a.user_id=$1 GROUP BY a.id, a.title, a.content, a.updated_at, a.created_at, u.id, u.name, u.email ORDER BY a.id DESC", userID)
	return articles, err
}

//UpdateArticle ...
func UpdateArticle(userID int64, id int64, form forms.ArticleForm) (err error) {
	_, err = GetArticle(userID, id)

	if err != nil {
		return errors.New("Article not found")
	}

	getDb := db.GetDB()

	_, err = getDb.Exec("UPDATE article SET title=$1, content=$2, updated_at=$3 WHERE id=$4", form.Title, form.Content, time.Now().Unix(), id)

	if err != nil {
		fmt.Printf("UpdateArticle Err: %v", err)
	}

	return err
}

//DeleteArticle ...
func DeleteArticle(userID int64, id int64) (err error) {
	_, err = GetArticle(userID, id)

	if err != nil {
		return errors.New("Article not found")
	}

	getDb := db.GetDB()

	_, err = getDb.Exec("DELETE FROM article WHERE id=$1", id)

	if err != nil {
		fmt.Printf("DeleteArticle Err: %v", err)
	}

	return err
}
