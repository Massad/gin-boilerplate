package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Massad/gin-boilerplate/config"
	"github.com/Massad/gin-boilerplate/controllers"
	"github.com/Massad/gin-boilerplate/db"
	_ "github.com/Massad/gin-boilerplate/docs"
	"github.com/Massad/gin-boilerplate/forms"
	pgdb "github.com/Massad/gin-boilerplate/internal/db"
	"github.com/Massad/gin-boilerplate/middleware"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Golang Gin Boilerplate
// @version         3.0
// @description     A RESTful API boilerplate with Gin Framework, PostgreSQL, Redis and JWT authentication
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT License
// @license.url   https://github.com/Massad/gin-boilerplate/blob/master/LICENSE

// @host      localhost:9001
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

// runCreateAdmin bootstraps the platform superadmin through the dedicated admin
// pool: no HTTP route is ever exposed for this bootstrap.
func runCreateAdmin(ctx context.Context, getenv func(string) string, email, password, name string) error {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return errors.New("error: -email and -password are required")
	}
	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = "Superadmin"
	}
	if err := pgdb.InitAdmin(ctx, getenv); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := (models.AdminModel{}).Create(email, string(hashed), displayName); err != nil {
		return err
	}
	fmt.Printf("Superadmin %s created\n", strings.ToLower(strings.TrimSpace(email)))
	pgdb.Close()
	return nil
}

// spaIndexPath is the built SPA shell served single-origin by this binary.
const spaIndexPath = "./app/dist/index.html"

// serveSPAIndex serves the built SPA shell with no-cache semantics, falling
// back to the demo template when the frontend build is absent.
func serveSPAIndex(c *gin.Context) {
	if _, err := os.Stat(spaIndexPath); err != nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ginBoilerplateVersion": "v3.0",
			"goVersion":             runtime.Version(),
		})
		return
	}
	c.Header("Cache-Control", "no-cache")
	c.File(spaIndexPath)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error: failed to load the env file")
	}

	// The superadmin bootstrap is a CLI flag, parsed before any HTTP runtime:
	// the process exits right after the creation.
	var createAdmin bool
	var adminEmail, adminPassword, adminName string
	flag.BoolVar(&createAdmin, "create-admin", false, "create or reset the platform superadmin and exit")
	flag.StringVar(&adminEmail, "email", "", "superadmin email")
	flag.StringVar(&adminPassword, "password", "", "superadmin password")
	flag.StringVar(&adminName, "name", "Superadmin", "superadmin display name")
	flag.Parse()

	if createAdmin {
		if err := runCreateAdmin(context.Background(), os.Getenv, adminEmail, adminPassword, adminName); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := config.ValidateSigningSecrets(os.Getenv); err != nil {
		log.Fatal(err)
	}

	if os.Getenv("ENV") == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	binding.Validator = new(forms.DefaultValidator)

	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	ctx := context.Background()
	if err := pgdb.InitTenant(ctx, os.Getenv); err != nil {
		log.Fatal(err)
	}
	if err := pgdb.InitAdmin(ctx, os.Getenv); err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()
	if err := db.InitRedis(); err != nil {
		log.Fatal(err)
	}

	v1 := r.Group("/v1")
	{
		/*** START USER ***/
		user := new(controllers.UserController)

		v1.POST("/user/login", user.Login)
		v1.POST("/user/register", user.Register)
		v1.GET("/user/logout", middleware.TokenAuth(), user.Logout)

		/*** START AUTH ***/
		auth := new(controllers.AuthController)

		v1.POST("/login", auth.Login)
		v1.POST("/token/refresh", auth.Refresh)

		/*** START Article ***/
		article := new(controllers.ArticleController)

		v1.POST("/article", middleware.TokenAuth(), article.Create)
		v1.GET("/articles", middleware.TokenAuth(), article.All)
		v1.GET("/article/:id", middleware.TokenAuth(), article.One)
		v1.PUT("/article/:id", middleware.TokenAuth(), article.Update)
		v1.DELETE("/article/:id", middleware.TokenAuth(), article.Delete)

		/*** START ADMIN ***/
		admin := new(controllers.AdminController)

		v1.POST("/admin/login", admin.Login)
		adminAPI := v1.Group("/admin", middleware.AdminAuth())
		{
			adminAPI.GET("/logout", admin.Logout)
			adminAPI.GET("/tenants", admin.Tenants)
			adminAPI.POST("/tenants/:id/status", admin.SetTenantStatus)
		}
	}

	if os.Getenv("ENV") != "PRODUCTION" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	r.LoadHTMLGlob("./public/html/*")

	r.Static("/public", "./public")

	// Hashed SPA assets are immutable: a new build ships new filenames.
	assets := r.Group("/assets", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Next()
	})
	assets.Static("/", "./app/dist/assets")

	// The SPA shell is served with no-cache semantics. When the frontend build
	// is absent, the demo template keeps the binary fully functional (the Go
	// CI smoke test relies on that fallback).
	r.GET("/", serveSPAIndex)

	// Single-origin SPA hosting fallback: only API paths and non-GET requests
	// keep the JSON 404, everything else serves the SPA shell.
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/v1") || c.Request.Method != http.MethodGet {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Not found"})
			return
		}
		serveSPAIndex(c)
	})

	port := os.Getenv("PORT")

	log.Printf("\n\n PORT: %s \n ENV: %s \n SSL: %s \n Version: %s \n\n", port, os.Getenv("ENV"), os.Getenv("SSL"), os.Getenv("API_VERSION"))

	if os.Getenv("SSL") == "TRUE" {
		SSLKeys := &struct {
			CERT string
			KEY  string
		}{
			CERT: "./cert/myCA.cer",
			KEY:  "./cert/myCA.key",
		}

		r.RunTLS(":"+port, SSLKeys.CERT, SSLKeys.KEY)
	} else {
		r.Run(":" + port)
	}
}
