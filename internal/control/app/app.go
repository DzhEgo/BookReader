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
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"os"
	"os/signal"
	"syscall"
)

type app struct {
	ctx      context.Context
	cancel   context.CancelFunc
	stopChan chan struct{}

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

	ctx, cancel := context.WithCancel(context.Background())
	a := &app{
		ctx:      ctx,
		cancel:   cancel,
		fApp:     newFiberApp(cfg),
		stopChan: make(chan struct{}),
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

	a.catchStop()
	return a.run()
}

func (a *app) catchStop() {
	go func() {
		exitChan := make(chan os.Signal, 10)
		signal.Notify(exitChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-exitChan:
			log.Info(fmt.Sprintf("%s caught interrupt, stopping...", sig.String()))
			a.stop()
		case <-a.ctx.Done():
			return
		}
	}()
}

func (a *app) stop() {
	a.cancel()
	a.srv.Cache.Clean()
	a.stopChan <- struct{}{}
}

func (a *app) run() error {
	a.listen()
	log.Info("application is running")

	<-a.stopChan
	log.Info("exit")
	return nil
}

func (a *app) listen() {
	go func() {
		if err := a.fApp.Listen(":8080"); err != nil {
			log.Error("API server error")
			a.stop()
		}
	}()
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
	srv.Cache = cache.NewService()
	srv.Reader = reader.NewService(
		reader.WithCache(srv.Cache),
	)
	srv.Books = books.NewService(
		books.WithCache(srv.Cache),
		books.WithReader(srv.Reader),
	)

	a.srv = srv
	return err
}
