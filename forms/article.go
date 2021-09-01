package forms

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

//ArticleForm ...
type ArticleForm struct{}

//CreateArticleForm ...
type CreateArticleForm struct {
	Title   string `form:"title" json:"title" binding:"required,min=3,max=100"`
	Content string `form:"content" json:"content" binding:"required,min=3,max=1000"`
}

//Title ...
func (f ArticleForm) Title(tag string, errMsg ...string) (message string) {
	switch tag {
	case "required":
		if len(errMsg) == 0 {
			return "Please enter the article title"
		}
		return errMsg[0]
	case "min", "max":
		return "Title should be between 3 to 100 characters"
	default:
		return "Something went wrong, please try again later"
	}
}

//Content ...
func (f ArticleForm) Content(tag string, errMsg ...string) (message string) {
	switch tag {
	case "required":
		if len(errMsg) == 0 {
			return "Please enter the article content"
		}
		return errMsg[0]
	case "min", "max":
		return "Content should be between 3 to 1000 characters"
	default:
		return "Something went wrong, please try again later"
	}
}

//Create ...
func (f ArticleForm) Create(err error) string {
	switch err.(type) {
	case validator.ValidationErrors:

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return "Something went wrong, please try again later"
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Title" {
				return f.Title(err.Tag())
			}
			if err.Field() == "Content" {
				return f.Content(err.Tag())
			}
		}

	default:
		return "Invalid request"
	}

	return "Something went wrong, please try again later"
}

//Update ...
func (f ArticleForm) Update(err error) string {
	switch err.(type) {
	case validator.ValidationErrors:

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return "Something went wrong, please try again later"
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Title" {
				return f.Title(err.Tag())
			}
			if err.Field() == "Content" {
				return f.Content(err.Tag())
			}
		}

	default:
		return "Invalid request"
	}

	return "Something went wrong, please try again later"
}
