// +build all

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Massad/gin-boilerplate/controllers"
	"github.com/Massad/gin-boilerplate/db"
	"github.com/Massad/gin-boilerplate/forms"

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
		user := new(controllers.UserController)

		v1.POST("/user/signin", user.Signin)
		v1.POST("/user/signup", user.Signup)
		v1.GET("/user/signout", user.Signout)

		/*** START Article ***/
		article := new(controllers.ArticleController)

		v1.POST("/article", article.Create)
		v1.GET("/articles", article.All)
		v1.GET("/article/:id", article.One)
		v1.PUT("/article/:id", article.Update)
		v1.DELETE("/article/:id", article.Delete)
	}

	return r
}

func main() {
	r := SetupRouter()
	r.Run()
}

var signinCookie string

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
* TestSignup
* Test user registration
*
* Must return response code 200
 */
func TestSignup(t *testing.T) {
	testRouter := SetupRouter()

	var signupForm forms.SignupForm

	signupForm.Name = "testing"
	signupForm.Email = testEmail
	signupForm.Password = testPassword

	data, _ := json.Marshal(signupForm)

	req, err := http.NewRequest("POST", "/v1/user/signup", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
}

/**
* TestSignupInvalidEmail
* Test user registration with invalid email
*
* Must return response code 406
 */
func TestSignupInvalidEmail(t *testing.T) {
	testRouter := SetupRouter()

	var signupForm forms.SignupForm

	signupForm.Name = "testing"
	signupForm.Email = "invalid@email"
	signupForm.Password = testPassword

	data, _ := json.Marshal(signupForm)

	req, err := http.NewRequest("POST", "/v1/user/signup", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 400) //406
}

/**
* TestSignin
* Test user signin
* and store the cookie on local variable [signinCookie]
*
* Must return response code 200
 */
func TestSignin(t *testing.T) {
	testRouter := SetupRouter()

	var signinForm forms.SigninForm

	signinForm.Email = testEmail
	signinForm.Password = testPassword

	data, _ := json.Marshal(signinForm)

	req, err := http.NewRequest("POST", "/v1/user/signin", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)

	signinCookie = resp.Header().Get("Set-Cookie")

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
	req.Header.Set("Cookie", signinCookie)

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
	req.Header.Set("Cookie", signinCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 400) //406
}

/**
* TestCreateArticleNotSignedIn
* Test article creation with a not signed in user
*
* Must return response code 403
 */
func TestCreateArticleNotSignedIn(t *testing.T) {
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
	req.Header.Set("Cookie", signinCookie)

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
	req.Header.Set("Cookie", signinCookie)

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
	req.Header.Set("Cookie", signinCookie)

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
	req.Header.Set("Cookie", signinCookie)

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
}

/**
* TestUserSignout
* Test signout a user
*
* Must return response code 200
 */
func TestUserSignout(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", "/v1/user/signout", nil)

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
