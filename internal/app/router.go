// Package app assembles the production Gin engine: middleware chain, /v1 API
// routes, Swagger (non-production) and single-origin SPA hosting. It is the
// composition root shared by the cmd/api entrypoint and the integration suite.
package app

import (
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/Massad/gin-boilerplate/controllers"
	_ "github.com/Massad/gin-boilerplate/docs"
	"github.com/Massad/gin-boilerplate/internal/modules/article"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	platformmw "github.com/Massad/gin-boilerplate/internal/platform/middleware"
	validate "github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/Massad/gin-boilerplate/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Deps carries the entrypoint inputs the router needs: the template glob for
// LoadHTMLGlob and the environment (the Swagger mount is dev-only).
type Deps struct {
	TemplateGlob string
	Env          string
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

// NewRouter builds the production engine. The wiring is copied verbatim from
// the former main.go blocks: gin.Default, trusted proxies, the global form
// validator, CORS/RequestID/gzip, the /v1 surface, Swagger and the SPA host.
func NewRouter(deps Deps) *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	binding.Validator = new(validate.DefaultValidator)

	r.Use(platformmw.CORS())
	r.Use(platformmw.RequestID())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

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
		articles := article.NewService(article.NewRepository(pgdb.TenantStore()))
		articleHandler := article.NewHandler(articles)

		v1.POST("/article", middleware.TokenAuth(), articleHandler.Create)
		v1.GET("/articles", middleware.TokenAuth(), articleHandler.All)
		v1.GET("/article/:id", middleware.TokenAuth(), articleHandler.One)
		v1.PUT("/article/:id", middleware.TokenAuth(), articleHandler.Update)
		v1.DELETE("/article/:id", middleware.TokenAuth(), articleHandler.Delete)

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

	if deps.Env != "PRODUCTION" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	r.LoadHTMLGlob(deps.TemplateGlob)

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

	return r
}
