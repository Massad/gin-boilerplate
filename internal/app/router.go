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
	"github.com/Massad/gin-boilerplate/internal/modules/admin"
	"github.com/Massad/gin-boilerplate/internal/modules/article"
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	user "github.com/Massad/gin-boilerplate/internal/modules/user"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	platformmw "github.com/Massad/gin-boilerplate/internal/platform/middleware"
	redisdb "github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	validate "github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Deps carries the entrypoint inputs the router needs: the template glob for
// LoadHTMLGlob, the environment (the Swagger mount is dev-only) and the three
// signing secrets of the auth service (validated upstream by the entrypoint).
type Deps struct {
	TemplateGlob string
	Env          string
	Secrets      auth.Secrets
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
		/*** START AUTH MODULE ***/
		// Composition root of the auth module: the Redis client is read from
		// the platform singleton here for the transitory window (dette tracée
		// pour post-Phase 8 : singletons prohibés hors composition root).
		authSvc := auth.NewService(deps.Secrets, auth.NewRepository(redisdb.GetRedis()))
		authHandler := auth.NewHandler(authSvc)
		authSvc.StartSweep(auth.DefaultSweepInterval)

		/*** START USER ***/
		// Composition root of the user module: the platform stores and the
		// tenant repository are read/resolved here for the transitory window
		// (dette tracée pour post-Phase 8 : singletons prohibés hors
		// composition root).
		tenants := tenant.NewRepository(pgdb.AdminStore())
		users := user.NewService(user.NewRepository(pgdb.AdminStore(), pgdb.TenantStore()), tenants, authSvc, pgdb.AdminStore())
		userHandler := user.NewHandler(users)

		v1.POST("/user/login", userHandler.Login)
		v1.POST("/user/register", userHandler.Register)
		v1.GET("/user/logout", auth.TokenAuth(authSvc), authHandler.UserLogout)

		/*** START ADMIN ***/
		// Composition root of the admin module: the platform store is
		// resolved here for the transitory window (dette tracée pour
		// post-Phase 8 : singletons prohibés hors composition root). The
		// section sits before the unified-auth block because the unified
		// login consumes the admin service in its superadmin fallback.
		admins := admin.NewService(admin.NewRepository(pgdb.AdminStore()), tenants, authSvc, pgdb.AdminStore())
		adminHandler := admin.NewHandler(admins, authSvc)

		v1.POST("/admin/login", adminHandler.Login)
		adminAPI := v1.Group("/admin", auth.AdminAuth(authSvc))
		{
			adminAPI.GET("/logout", authHandler.AdminLogout)
			adminAPI.GET("/tenants", adminHandler.Tenants)
			adminAPI.POST("/tenants/:id/status", adminHandler.SetTenantStatus)
		}

		/*** START AUTH ***/
		authController := controllers.NewAuthController(authSvc, users, admins)

		v1.POST("/login", authController.Login)
		v1.POST("/token/refresh", authController.Refresh)

		/*** START Article ***/
		articles := article.NewService(article.NewRepository(pgdb.TenantStore()))
		articleHandler := article.NewHandler(articles)

		v1.POST("/article", auth.TokenAuth(authSvc), articleHandler.Create)
		v1.GET("/articles", auth.TokenAuth(authSvc), articleHandler.All)
		v1.GET("/article/:id", auth.TokenAuth(authSvc), articleHandler.One)
		v1.PUT("/article/:id", auth.TokenAuth(authSvc), articleHandler.Update)
		v1.DELETE("/article/:id", auth.TokenAuth(authSvc), articleHandler.Delete)
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
