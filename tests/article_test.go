package tests

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gin-boilerplate/controllers"
	"gin-boilerplate/db"
	"gin-boilerplate/forms"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	gin.SetMode(gin.TestMode)

	store, _ := sessions.NewRedisStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	r.Use(sessions.Sessions("gin-boilerplate-session", store))

	v1 := r.Group("/v1")
	{
		/*** START USER ***/
		v1.POST("/user/login", controllers.Login)
		v1.POST("/user/register", controllers.Register)
		v1.GET("/user/logout", controllers.Logout)

		/*** START Article ***/
		v1.POST("/article", controllers.CreateArticle)
		v1.GET("/article/:id", controllers.GetArticle)
		v1.GET("/articles", controllers.GetArticles)
		v1.PUT("/article/:id", controllers.UpdateArticle)
		v1.DELETE("/article/:id", controllers.DeleteArticle)
	}

	return r
}

func main() {
	r := SetupRouter()
	r.Run()
}

var loginCookie string

var testEmail = "test-gin-boilerplate@test.com"
var testPassword = "123456"

var articleID int

/**
* TestIntDB
* It tests the connection to the database and init the db for this test
*
* Must pass
 */
func TestIntDB(t *testing.T) {
	db.Init()
}

/**
* TestRegister
* Test user registration
*
* Must return response code 200
 */
func TestRegister(t *testing.T) {
	testRouter := SetupRouter()

	var registerForm forms.RegisterForm

	registerForm.Name = "testing"
	registerForm.Email = testEmail
	registerForm.Password = testPassword

	data, _ := json.Marshal(registerForm)

	req, err := http.NewRequest("POST", "/v1/user/register", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
}

/**
* TestRegisterInvalidEmail
* Test user registration with invalid email
*
* Must return response code 406
 */
func TestRegisterInvalidEmail(t *testing.T) {
	testRouter := SetupRouter()

	var registerForm forms.RegisterForm

	registerForm.Name = "testing"
	registerForm.Email = "invalid@email"
	registerForm.Password = testPassword

	data, _ := json.Marshal(registerForm)

	req, err := http.NewRequest("POST", "/v1/user/register", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 406)
}

/**
* TestLogin
* Test user login
* and store the cookie on local variable [loginCookie]
*
* Must return response code 200
 */
func TestLogin(t *testing.T) {
	testRouter := SetupRouter()

	var loginForm forms.LoginForm

	loginForm.Email = testEmail
	loginForm.Password = testPassword

	data, _ := json.Marshal(loginForm)

	req, err := http.NewRequest("POST", "/v1/user/login", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)

	loginCookie = resp.Header().Get("Set-Cookie")

	assert.Equal(t, resp.Code, 200)
}

/**
* TestCreateArticle
* Test article creation
*
* Must return response code 200
 */
func TestCreateArticle(t *testing.T) {
	testRouter := SetupRouter()

	var articleForm forms.ArticleForm

	articleForm.Title = "Testing article title"
	articleForm.Content = "Testing article content"

	data, _ := json.Marshal(articleForm)

	req, err := http.NewRequest("POST", "/v1/article", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", loginCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	res := struct {
		Status int
		ID     int
	}{}

	json.Unmarshal(body, &res)

	articleID = res.ID

	assert.Equal(t, resp.Code, 200)
}

/**
* TestCreateInvalidArticle
* Test article invalid creation
*
* Must return response code 406
 */
func TestCreateInvalidArticle(t *testing.T) {
	testRouter := SetupRouter()

	var articleForm forms.ArticleForm

	articleForm.Title = "Testing article title"

	data, _ := json.Marshal(articleForm)

	req, err := http.NewRequest("POST", "/v1/article", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", loginCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 406)
}

/**
* TestCreateArticleNotLoggedIn
* Test article creation with a not logged in user
*
* Must return response code 403
 */
func TestCreateArticleNotLoggedIn(t *testing.T) {
	testRouter := SetupRouter()

	var articleForm forms.ArticleForm

	articleForm.Title = "Testing article title"
	articleForm.Content = "Testing article content"

	data, _ := json.Marshal(articleForm)

	req, err := http.NewRequest("POST", "/v1/article", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 403)
}

/**
* TestGetArticle
* Test getting one article
*
* Must return response code 200
 */
func TestGetArticle(t *testing.T) {
	testRouter := SetupRouter()

	url := fmt.Sprintf("/v1/article/%d", articleID)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", loginCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
}

/**
* TestGetInvalidArticle
* Test getting invalid article
*
* Must return response code 404
 */
func TestGetInvalidArticle(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", "/v1/article/invalid", nil)
	req.Header.Set("Cookie", loginCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

/**
* TestUpdateArticle
* Test updating an article
*
* Must return response code 200
 */
func TestUpdateArticle(t *testing.T) {
	testRouter := SetupRouter()

	var articleForm forms.ArticleForm

	articleForm.Title = "Testing new article title"
	articleForm.Content = "Testing new article content"

	data, _ := json.Marshal(articleForm)

	url := fmt.Sprintf("/v1/article/%d", articleID)

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", loginCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
}

/**
* TestDeleteArticle
* Test deleting an article
*
* Must return response code 200
 */
func TestDeleteArticle(t *testing.T) {
	testRouter := SetupRouter()

	url := fmt.Sprintf("/v1/article/%d", articleID)

	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Cookie", loginCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
}

/**
* TestUserLogout
* Test logout a user
*
* Must return response code 200
 */
func TestUserLogout(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", "/v1/user/logout", nil)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
}

/**
* TestCleanUp
* Deletes the created user with it's articles
*
* Must pass
 */
func TestCleanUp(t *testing.T) {
	var err error
	_, err = db.GetDB().Exec("DELETE FROM public.user WHERE email=$1", testEmail)
	if err != nil {
		t.Error(err)
	}
}
