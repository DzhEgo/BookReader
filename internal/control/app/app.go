package app

import (
	"BookStore/docs"
	"BookStore/internal/common/config"
	"BookStore/internal/control/api"
	"BookStore/internal/control/service"
	"BookStore/internal/control/service/auth"
	"BookStore/internal/control/service/books"
	"BookStore/internal/control/service/cache"
	"BookStore/internal/control/service/reader"
	"BookStore/internal/control/service/users"
	"BookStore/internal/database"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

type app struct {
	fApp *fiber.App
	api  *api.ApiHandler
	srv  service.Services
}

//	@title			Swagger Book API
//	@version		1.0
//	@description	This is a sample server celler server.
//	@termsOfService	http://swagger.io/terms/

// @host		localhost:8080
// @BasePath	/api/v1
func Run(cfg *config.BaseConfig) (err error) {
	a := &app{
		fApp: newFiberApp(cfg),
	}

	if err = godotenv.Load(); err != nil {
		log.Fatal(".env file not found")
	}

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}

	if err := a.initServices(); err != nil {
		return fmt.Errorf("init services: %w", err)
	}

	a.api, err = api.NewApiHandler(cfg, a.fApp.Group("/api/v1"), a.srv)
	if err != nil {
		return err
	}

	a.fApp.Static("/", "./front")
	a.fApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./front/index.html")
	})

	log.Info("app started")

	return a.fApp.Listen(":8080")
}

func newFiberApp(cfg *config.BaseConfig) *fiber.App {
	app := fiber.New()
	if cfg.Swagger {
		docs.SwaggerInfo.Title = "Swagger BookStore API"
		docs.SwaggerInfo.Version = "1.0"
		docs.SwaggerInfo.Host = fmt.Sprintf("0.0.0.0:8080")
		docs.SwaggerInfo.BasePath = "/api/v1"
		app.Get("/swagger/*", fiberSwagger.WrapHandler)
	}
	return app
}

func (a *app) initServices() (err error) {
	var srv service.Services
	srv.Auth = auth.NewService()
	srv.User = users.NewService()
	cacheService := cache.NewService()
	srv.Cache = cacheService
	readerService := reader.NewService(cacheService)
	srv.Reader = readerService
	srv.Books = books.NewService(readerService, cacheService)

	a.srv = srv
	return err
}
