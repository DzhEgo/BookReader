package api

import (
	"BookStore/internal/common/config"
	"BookStore/internal/common/utils"
	"BookStore/internal/control/service"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

type ApiHandler struct {
	cfg    *config.BaseConfig
	router fiber.Router
	routes [][]*fiber.Route
	srv    service.Services
	secret [32]byte
}

type Option func(ah *ApiHandler)

func NewApiHandler(
	cfg *config.BaseConfig,
	router fiber.Router,
	srv service.Services,
	opts ...Option,
) (*ApiHandler, error) {
	ah := &ApiHandler{
		cfg:    cfg,
		router: router,
		srv:    srv,
	}

	for _, opt := range opts {
		opt(ah)
	}

	n, err := rand.Read(ah.secret[:])
	if err != nil {
		return nil, fmt.Errorf("secret key: %w", err)
	}
	if n != len(ah.secret) {
		return nil, errors.New("partially empty secret key")
	}

	b := ah.router.Group("/book")
	b.Get("/list", ah.getBook)
	ah.router.Post("/registration", ah.registration)
	ah.router.Post("/login", ah.login)

	ah.router.Use(
		jwtware.New(jwtware.Config{
			SigningKey:  ah.secret[:],
			TokenLookup: "header:Authorization",
			AuthScheme:  "Bearer",
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				code := fiber.StatusUnauthorized
				var e *fiber.Error
				if errors.As(err, &e) {
					code = e.Code
				}
				return utils.Response(ctx, code, "authorization fail")
			},
		}),
	)

	ah.router.Post("/logout", ah.logout)
	ah.router.Get("/profile", ah.profile)

	admin := ah.router.Group("/admin", ah.roleMiddleware("admin"))
	admin.Get("/user/list", ah.users)
	admin.Put("/role/set", ah.updateRole)
	admin.Get("/role/list", ah.roles)
	admin.Delete("/user/delete", ah.deleteUser)

	b.Post("/upload", ah.uploadBook)
	b.Delete("/delete", ah.deleteBook)
	b.Get("/read", ah.getBookPage)
	b.Post("/progress/set", ah.saveProgress)
	b.Get("/progress/get", ah.getProgress)

	return ah, nil
}
