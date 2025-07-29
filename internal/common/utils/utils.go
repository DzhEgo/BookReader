package utils

import (
	"BookStore/internal/common/model"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func Response(ctx *fiber.Ctx, code int, message string) error {
	r := &model.Response{Code: code, Message: message}
	return ctx.Status(code).JSON(r)
}

func Responsef(ctx *fiber.Ctx, code int, message string, params ...any) error {
	r := &model.Response{Code: code, Message: fmt.Sprintf(message, params...)}
	return ctx.Status(code).JSON(r)
}
