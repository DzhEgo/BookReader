package api

import (
	"BookStore/internal/common/utils"
	"BookStore/internal/control/model"
	dbmodel "BookStore/internal/database/model"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// @Summary	create user
// @ID			createUser
// @Accept		json
// @Param		params	body		model.Creditionals	true	"Creditionals's credentials"	request
// @Failure	500		{object}	model.Response		"Internal Server Error"
// @Failure	400		{object}	model.Response		"Bad Request"
// @Failure	401		{object}	model.Response		"Unauthorized"
// @Success	200		{object}	string				"OK"
// @Router		/registration [post]
func (ah *ApiHandler) registration(ctx *fiber.Ctx) error {
	var cred model.Creditionals

	if err := ctx.BodyParser(&cred); err != nil {
		log.Errorf("failed to unmarshal json: %v", err)
		wrapErr := fmt.Errorf("failed to unmarshal json: %v", err)
		return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
	}

	if cred.Login == "" && cred.Password == "" {
		log.Errorf("login or password is required")
		wrapErr := fmt.Errorf("login or password is required")
		return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
	}

	err := ah.srv.Auth.CreateUser(cred)
	if err != nil {
		log.Errorf("failed to create user: %v", err)
		wrapErr := fmt.Errorf("failed to create user: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return utils.Response(ctx, fiber.StatusOK, "OK")
}

// @Summary	login
// @ID			login
// @Accept		json
// @Param		params	body		model.Creditionals	true	"Creditionals's credentials"	request
// @Failure	500		{object}	model.Response		"Internal Server Error"
// @Failure	400		{object}	model.Response		"Bad Request"
// @Failure	401		{object}	model.Response		"Unauthorized"
// @Success	200		{object}	string				"OK"
// @Router		/login [post]
func (ah *ApiHandler) login(ctx *fiber.Ctx) error {
	var cred model.Creditionals
	var ok bool

	if err := ctx.BodyParser(&cred); err != nil {
		log.Errorf("failed to unmarshal json: %v", err)
		wrapErr := fmt.Errorf("failed to unmarshal json: %v", err)
		return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
	}

	if cred.Login == "" || cred.Password == "" {
		wrapErr := fmt.Errorf("invalid login or password")
		return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
	}

	user, err := ah.getUser(ctx, cred)
	if err != nil {
		log.Errorf("failed to get user: %v", err)
		wrapErr := fmt.Errorf("failed to get user: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	ok, _ = ah.srv.Auth.ValidateUser(user, cred)
	if !ok {
		wrapErr := fmt.Errorf("invalid user")
		return utils.Response(ctx, fiber.StatusUnauthorized, wrapErr.Error())
	}

	now := time.Now()
	token, err := model.NewToken(ah.secret[:], user.Login, user.Role.RoleName, now)
	if err != nil {
		log.Errorf("failed to create token: %v", err)
		wrapErr := fmt.Errorf("failed to create token: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}
	ah.setCookies(ctx, token, now)

	return ctx.JSON(token)
}

func (ah *ApiHandler) getUser(ctx *fiber.Ctx, cred model.Creditionals) (*dbmodel.User, error) {
	login := model.Creditionals{Login: cred.Login}

	user, err := ah.srv.Auth.GetUser(login)
	if err != nil {
		log.Errorf("failed to get user: %v", err)
		wrapErr := fmt.Errorf("failed to get user: %v", err)
		return nil, utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return user, nil
}

// @Summary	logout
// @ID			logout
// @Accept		json
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	string			"OK"
// @Router		/logout [post]
func (ah *ApiHandler) logout(ctx *fiber.Ctx) error {
	ctx.ClearCookie()
	ah.clearCookies(ctx, time.Now())

	return utils.Response(ctx, fiber.StatusOK, "OK")
}

// @Summary	refresh
// @ID			refresh
// @Accept		json
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	string			"OK"
// @Router		/refresh [get]
func (ah *ApiHandler) refresh(ctx *fiber.Ctx) error {
	tokenStr := ctx.Cookies("refresh_token")
	claims := make(jwt.MapClaims, 2)

	_, err := jwt.ParseWithClaims(
		tokenStr, claims,
		func(token *jwt.Token) (interface{}, error) {
			return ah.secret[:], nil
		},
	)
	if err != nil {
		return utils.Response(ctx, fiber.StatusUnauthorized, "invalid token")
	}

	refreshVal, ok := claims["refresh"]
	if !ok {
		return utils.Response(ctx, fiber.StatusUnauthorized, "refresh field missing")
	}

	refresh, ok := refreshVal.(bool)
	if !ok {
		return utils.Response(ctx, fiber.StatusUnauthorized, "invalid token")
	}

	if !refresh {
		return utils.Response(ctx, fiber.StatusUnauthorized, "invalid token")
	}

	login, ok := claims["login"].(string)
	if !ok {
		return utils.Response(ctx, fiber.StatusUnauthorized, "invalid token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return utils.Response(ctx, fiber.StatusUnauthorized, "invalid token")
	}

	now := time.Now()
	token, err := model.NewToken(ah.secret[:], login, role, now)
	if err != nil {
		log.Errorf("failed to create token: %v", err)
		wrapErr := fmt.Errorf("failed to create token: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}
	ah.setCookies(ctx, token, now)

	return ctx.JSON(token)
}

func (ah *ApiHandler) setCookies(ctx *fiber.Ctx, token *model.Token, createTime time.Time) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token.Token,
		Expires:  createTime.Add(72 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "none",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token.RefreshToken,
		Expires:  createTime.Add(25 * 365 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "none",
	})
}

func (ah *ApiHandler) clearCookies(ctx *fiber.Ctx, createTime time.Time) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  createTime,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "none",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  createTime,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "none",
	})
}

func (ah *ApiHandler) roleMiddleware(required ...string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userToken := ctx.Locals("user")
		if userToken == nil {
			return utils.Response(ctx, fiber.StatusUnauthorized, "unauthorized")
		}

		user, ok := userToken.(*jwt.Token)
		if !ok {
			return utils.Response(ctx, fiber.StatusUnauthorized, "invalid token format")
		}
		claims, ok := user.Claims.(jwt.MapClaims)
		if !ok {
			return utils.Response(ctx, fiber.StatusUnauthorized, "invalid token format")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return utils.Response(ctx, fiber.StatusUnauthorized, "invalid role")
		}

		for _, k := range required {
			if role == k {
				return ctx.Next()
			}
		}
		return utils.Response(ctx, fiber.StatusForbidden, "insufficient permissions")
	}

}

func (ah *ApiHandler) getUserFromContext(ctx *fiber.Ctx) (*model.UserContext, error) {
	userToken := ctx.Locals("user")
	if userToken == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	user, ok := userToken.(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("invalid token format")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token format")
	}

	login, ok := claims["login"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token format")
	}

	return ah.srv.Auth.GetUserContext(model.Creditionals{Login: login})
}
