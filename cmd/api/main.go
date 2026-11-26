package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Massad/gin-boilerplate/internal/app"
	"github.com/Massad/gin-boilerplate/internal/modules/admin"
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/platform/config"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	// The bootstrap drives the superadmin upsert only: it never issues
	// tokens, lists tenants or opens a unit of work (nil ports — covered by
	// cmd/api/main_test.go).
	admins := admin.NewService(admin.NewRepository(pgdb.AdminStore()), nil, nil, nil)
	if _, err := admins.Create(email, string(hashed), displayName); err != nil {
		return err
	}
	fmt.Printf("Superadmin %s created\n", strings.ToLower(strings.TrimSpace(email)))
	pgdb.Close()
	return nil
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

	r := app.NewRouter(app.Deps{
		TemplateGlob: "./public/html/*",
		Env:          os.Getenv("ENV"),
		Secrets: auth.Secrets{
			Access:      os.Getenv("ACCESS_SECRET"),
			Refresh:     os.Getenv("REFRESH_SECRET"),
			AdminAccess: os.Getenv("ADMIN_ACCESS_SECRET"),
		},
	})

	ctx := context.Background()
	if err := pgdb.InitTenant(ctx, os.Getenv); err != nil {
		log.Fatal(err)
	}
	if err := pgdb.InitAdmin(ctx, os.Getenv); err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()
	if err := redisdb.InitRedis(); err != nil {
		log.Fatal(err)
	}

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
