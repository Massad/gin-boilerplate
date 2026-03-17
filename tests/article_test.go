//go:build all
// +build all

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Massad/gin-boilerplate/controllers"
	"github.com/Massad/gin-boilerplate/db"
	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/invoice"
	"github.com/Massad/gin-boilerplate/middleware"
	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	gin.SetMode(gin.TestMode)

	//Custom form validator
	binding.Validator = new(forms.DefaultValidator)

	v1 := r.Group("/v1")
	{
		/*** START USER ***/
		user := new(controllers.UserController)

		v1.POST("/user/login", user.Login)
		v1.POST("/user/register", user.Register)
		v1.GET("/user/logout", middleware.TokenAuth(), user.Logout)

		/*** START AUTH ***/
		auth := new(controllers.AuthController)

		v1.POST("/token/refresh", auth.Refresh)

		/*** START Article ***/
		article := new(controllers.ArticleController)

		v1.POST("/article", middleware.TokenAuth(), article.Create)
		v1.GET("/articles", middleware.TokenAuth(), article.All)
		v1.GET("/article/:id", middleware.TokenAuth(), article.One)
		v1.PUT("/article/:id", middleware.TokenAuth(), article.Update)
		v1.DELETE("/article/:id", middleware.TokenAuth(), article.Delete)

		/*** START Invoice ***/
		inv := new(invoice.InvoiceController)

		v1.GET("/invoice", inv.Preview)
		v1.GET("/invoice/download", inv.Download)
	}

	r.LoadHTMLGlob("../public/html/*")

	return r
}

func main() {
	r := SetupRouter()
	r.Run()
}

var loginCookie string

var testEmail = "test-gin-boilerplate@test.com"
var testPassword = "123456"

var accessToken string
var refreshToken string

var articleID int

/**
* TestIntDB
* It tests the connection to the database and init the db for this test
*
* Must pass
 */
func TestIntDB(t *testing.T) {

	//Load the .env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file, please create one in the root directory")
	}

	db.Init()
	db.InitRedis(1)
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
	assert.Equal(t, http.StatusOK, resp.Code)
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
	assert.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestLogin
* Test user login
* and get the access_token and refresh_token stored
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var res struct {
		Message string `json:"message"`
		User    struct {
			CreatedAt int64  `json:"created_at"`
			Email     string `json:"email"`
			ID        int64  `json:"id"`
			Name      string `json:"name"`
			UpdatedAt int64  `json:"updated_at"`
		} `json:"user"`
		Token struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"token"`
	}
	json.Unmarshal(body, &res)

	accessToken = res.Token.AccessToken
	refreshToken = res.Token.RefreshToken

	assert.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestInvalidLogin
* Test invalid login
*
* Must return response code 406
 */
func TestInvalidLogin(t *testing.T) {
	testRouter := SetupRouter()

	var loginForm forms.LoginForm

	loginForm.Email = "wrong@email.com"
	loginForm.Password = testPassword

	data, _ := json.Marshal(loginForm)

	req, err := http.NewRequest("POST", "/v1/user/login", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()

	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestCreateArticle
* Test article creation
*
* Must return response code 200
 */
func TestCreateArticle(t *testing.T) {
	testRouter := SetupRouter()

	var form forms.CreateArticleForm

	form.Title = "Testing article title"
	form.Content = "Testing article content"

	data, _ := json.Marshal(form)

	req, err := http.NewRequest("POST", "/v1/article", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessToken))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var res struct {
		Status int
		ID     int
	}
	json.Unmarshal(body, &res)

	articleID = res.ID

	assert.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestCreateInvalidArticle
* Test article invalid creation
*
* Must return response code 406
 */
func TestCreateInvalidArticle(t *testing.T) {
	testRouter := SetupRouter()

	var form forms.CreateArticleForm

	form.Title = "Testing article title"

	data, _ := json.Marshal(form)

	req, err := http.NewRequest("POST", "/v1/article", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessToken))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestGetArticle
* Test getting one article
*
* Must return response code 200
 */
func TestGetArticle(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", fmt.Sprintf("/v1/article/%d", articleID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessToken))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessToken))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

/**
* TestGetArticleNotLoggedin
* Test getting the article with logged out user
*
* Must return response code 401
 */
func TestGetArticleNotLoggedin(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", fmt.Sprintf("/v1/article/%d", articleID), nil)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestGetArticleUnauthorized
* Test getting the article with unauthorized user (wrong or expired access_token)
*
* Must return response code 401
 */
func TestGetArticleUnauthorized(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", fmt.Sprintf("/v1/article/%d", articleID), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", "abc123"))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestUpdateArticle
* Test updating an article
*
* Must return response code 200
 */
func TestUpdateArticle(t *testing.T) {
	testRouter := SetupRouter()

	var form forms.CreateArticleForm

	form.Title = "Testing new article title"
	form.Content = "Testing new article content"

	data, _ := json.Marshal(form)

	url := fmt.Sprintf("/v1/article/%d", articleID)

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessToken))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessToken))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestRefreshToken
* Test refreshing the token with valid refresh_token
*
* Must return response code 200
 */
func TestRefreshToken(t *testing.T) {
	testRouter := SetupRouter()

	var tokenForm forms.Token

	tokenForm.RefreshToken = refreshToken

	data, _ := json.Marshal(tokenForm)

	req, err := http.NewRequest("POST", "/v1/token/refresh", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestInvalidRefreshToken
* Test refreshing the token with invalid refresh_token
*
* Must return response code 401
 */
func TestInvalidRefreshToken(t *testing.T) {
	testRouter := SetupRouter()

	var tokenForm forms.Token

	//Since we didn't update it in the test before - this will not be valid anymore
	tokenForm.RefreshToken = refreshToken

	data, _ := json.Marshal(tokenForm)

	req, err := http.NewRequest("POST", "/v1/token/refresh", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestUserSignout
* Test logout a user
*
* Must return response code 200
 */
func TestUserLogout(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", "/v1/user/logout", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessToken))

	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestInvoicePreview
* Test invoice HTML preview
*
* Must return response code 200
 */
func TestInvoicePreview(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", "/v1/invoice", nil)
	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, resp.Body.String(), "INV-001")
	assert.Contains(t, resp.Body.String(), "Acme Corp")
}

/**
* TestInvoiceDownload
* Test invoice PDF download
*
* Must return response code 200 with PDF content type
 */
func TestInvoiceDownload(t *testing.T) {
	testRouter := SetupRouter()

	req, err := http.NewRequest("GET", "/v1/invoice/download", nil)
	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/pdf", resp.Header().Get("Content-Type"))
	assert.Contains(t, resp.Header().Get("Content-Disposition"), "INV-001.pdf")
	assert.True(t, resp.Body.Len() > 0, "PDF body should not be empty")
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
