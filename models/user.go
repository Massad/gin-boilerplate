package models

import (
	"errors"
	"gin-boilerplate/db"
	"gin-boilerplate/forms"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//User ...
type User struct {
	ID        int    `db:"id, primarykey, autoincrement" json:"id"`
	Email     string `db:"email" json:"email"`
	Password  string `db:"password" json:"-"`
	Name      string `db:"name" json:"name"`
	UpdatedAt int64  `db:"updated_at" json:"updated_at"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

//Login ...
func Login(form forms.LoginForm) (user User, err error) {

	err = db.GetDB().SelectOne(&user, "SELECT id, email, password, name, updated_at, created_at FROM public.user WHERE email=LOWER($1) LIMIT 1", form.Email)

	if err != nil {
		return user, err
	}

	bytePassword := []byte(form.Password)
	byteHashedPassword := []byte(user.Password)

	err = bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)

	if err != nil {
		return user, errors.New("Invalid password")
	}

	return user, nil
}

//Register ...
func Register(form forms.RegisterForm) (user User, err error) {
	getDb := db.GetDB()

	if !validateEmail(form.Email) {
		return user, errors.New("Email address is invalid")
	}

	checkUser, err := getDb.SelectInt("SELECT count(id) FROM public.user WHERE email=LOWER($1) LIMIT 1", form.Email)

	if err != nil {
		return user, err
	}

	if checkUser > 0 {
		return user, errors.New("User exists")
	}

	bytePassword := []byte(form.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	res, err := getDb.Exec("INSERT INTO public.user(email, password, name, updated_at, created_at) VALUES($1, $2, $3, $4, $5) RETURNING id", form.Email, string(hashedPassword), form.Name, time.Now().Unix(), time.Now().Unix())

	if res != nil && err == nil {
		err = getDb.SelectOne(&user, "SELECT id, email, name, updated_at, created_at FROM public.user WHERE email=LOWER($1) LIMIT 1", form.Email)
		if err == nil {
			return user, nil
		}
	}

	return user, errors.New("Not registered")
}

//GetUsers ...
func GetUsers() (u []User, err error) {
	var users []User
	_, err = db.GetDB().Select(&users, "SELECT * from public.user")
	return users, err
}

//CheckUser ...
func CheckUser(userID int64) (err error) {
	var user User
	err = db.GetDB().SelectOne(&user, "SELECT id FROM public.user WHERE id=$1", userID)
	return err
}
