package models

import (
	"errors"
	"fmt"
	"gin-boilerplate/db"
	"gin-boilerplate/forms"
	"time"
)

//Article ....
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
func CreateArticle(userID int64, form forms.ArticleForm) (err error) {
	getDb := db.GetDB()

	checkUser := CheckUser(userID)

	if checkUser != nil {
		return errors.New("User doesn't exist")
	}

	res, err := getDb.Exec("INSERT INTO article(user_id, title, content, updated_at, created_at) VALUES($1, $2, $3, $4, $5) RETURNING id", userID, form.Title, form.Content, time.Now().Unix(), time.Now().Unix())

	if res != nil && err == nil {
		return nil
	}

	return err
}

//GetArticle ...
func GetArticle(userID, id int64) (article Article) {
	var err error
	err = db.GetDB().SelectOne(&article, "SELECT a.id, a.title, a.content, a.updated_at, a.created_at, json_build_object('id', u.id, 'name', u.name, 'email', u.email) AS user FROM article a LEFT JOIN public.user u ON a.user_id = u.id WHERE a.user_id=$1 AND a.id=$2 GROUP BY a.id, a.title, a.content, a.updated_at, a.created_at, u.id, u.name, u.email LIMIT 1", userID, id)

	if err != nil {
		fmt.Printf("GetArticle Err: %v", err)
	}

	return article
}

//GetArticles ...
func GetArticles(userID int64) []Article {
	var articles []Article
	var err error
	_, err = db.GetDB().Select(&articles, "SELECT a.id, a.title, a.content, a.updated_at, a.created_at, json_build_object('id', u.id, 'name', u.name, 'email', u.email) AS user FROM article a LEFT JOIN public.user u ON a.user_id = u.id WHERE a.user_id=$1 GROUP BY a.id, a.title, a.content, a.updated_at, a.created_at, u.id, u.name, u.email ORDER BY a.id DESC", userID)

	if err != nil {
		fmt.Printf("GetArticles Err: %v", err)
	}

	return articles
}

//UpdateArticle ...
func UpdateArticle(userID int64, id int64, form forms.ArticleForm) (err error) {
	getDb := db.GetDB()

	checkArticle := GetArticle(userID, id)

	if checkArticle.ID == 0 {
		return errors.New("Article not found")
	}

	_, err = getDb.Exec("UPDATE article SET title=$1, content=$2, updated_at=$3 WHERE id=$4", form.Title, form.Content, time.Now().Unix(), id)

	if err != nil {
		fmt.Printf("UpdateArticle Err: %v", err)
	}

	return err
}

//DeleteArticle ...
func DeleteArticle(userID int64, id int64) (err error) {
	getDb := db.GetDB()

	checkArticle := GetArticle(userID, id)

	if checkArticle.ID == 0 {
		return errors.New("Article not found")
	}

	_, err = getDb.Exec("DELETE FROM article WHERE id=$1", id)

	if err != nil {
		fmt.Printf("DeleteArticle Err: %v", err)
	}

	return err
}
