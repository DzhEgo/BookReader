package api

import (
	"BookStore/internal/common/utils"
	"BookStore/internal/control/model"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"strconv"
)

// @Summary	get users
// @ID			getUser
// @Accept		json
// @Param		id	query		int				true	"User id"	request
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	model.User		"Data"
// @Router		/admin/user/list [get]
func (ah *ApiHandler) users(ctx *fiber.Ctx) error {
	idStr := ctx.Query("id")

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Errorf("failed to parse user id: %v", idStr)
			wrapErr := fmt.Errorf("failed to parse user id: %v", idStr)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}

		user, err := ah.srv.User.GetUser(id)
		if err != nil {
			log.Errorf("failed to get user: %v", err)
			wrapErr := fmt.Errorf("failed to get user: %v", err)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}
		return ctx.JSON(user)
	}

	users, err := ah.srv.User.Users()
	if err != nil {
		log.Errorf("failed to get users: %v", err)
		wrapErr := fmt.Errorf("failed to get users: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return ctx.JSON(users)
}

// @Summary	get roles
// @ID			getRole
// @Accept		json
// @Param		id	query		int				true	"Role id"	request
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	model.Role		"Data"
// @Router		/admin/role/list [get]
func (ah *ApiHandler) roles(ctx *fiber.Ctx) error {
	idStr := ctx.Query("id")
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Errorf("failed to parse role id: %v", idStr)
			wrapErr := fmt.Errorf("failed to parse role id: %v", idStr)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}

		role, err := ah.srv.User.GetRole(id)
		if err != nil {
			log.Errorf("failed to get role: %v", err)
			wrapErr := fmt.Errorf("failed to get role: %v", err)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}

		return ctx.JSON(role)
	}

	roles, err := ah.srv.User.GetRoles()
	if err != nil {
		log.Errorf("failed to get roles: %v", err)
		wrapErr := fmt.Errorf("failed to get roles: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return ctx.JSON(roles)
}

// @Summary	delete user
// @ID			deleteUser
// @Accept		json
// @Param		id	query		int				true	"User id"	request
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	string			"OK"
// @Router		/admin/user/delete [delete]
func (ah *ApiHandler) deleteUser(ctx *fiber.Ctx) error {
	idStr := ctx.Query("id")
	if idStr == "" {
		log.Errorf("failed to get user id: %v", idStr)
		wrapErr := fmt.Errorf("failed to get user id: %v", idStr)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Errorf("failed to parse user id: %v", idStr)
		wrapErr := fmt.Errorf("failed to parse user id: %v", idStr)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	userContext, err := ah.getUserFromContext(ctx)
	if err != nil {
		log.Errorf("failed to get user context: %v", err)
		wrapErr := fmt.Errorf("failed to get user context: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	if userContext.ID == id {
		log.Debugf("can not delete yourself")
		wrapErr := fmt.Errorf("cannot delete yourself")
		return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
	}

	if err := ah.srv.User.DeleteUser(id); err != nil {
		log.Errorf("failed to delete user %v: %v", id, err)
		wrapErr := fmt.Errorf("failed to delete user %v: %v", id, err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return nil
}

// @Summary	add role
// @ID			addRole
// @Accept		json
// @Param		params	body		model.AddRole	true	"Creditionals's credentials"	request
// @Failure	500		{object}	model.Response	"Internal Server Error"
// @Failure	400		{object}	model.Response	"Bad Request"
// @Failure	401		{object}	model.Response	"Unauthorized"
// @Success	200		{object}	string			"OK"
// @Router		/admin/addRole [post]
func (ah *ApiHandler) addRole(ctx *fiber.Ctx) error {
	var cred model.AddRole
	if err := ctx.BodyParser(&cred); err != nil {
		log.Errorf("failed to unmarshal json: %v", err)
		wrapErr := fmt.Errorf("failed to unmarshal json: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	userContext, err := ah.getUserFromContext(ctx)
	if err != nil {
		log.Errorf("failed to get user context: %v", err)
		wrapErr := fmt.Errorf("failed to get user context: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	if userContext.ID == cred.UserId {
		log.Debugf("can not change your role")
		wrapErr := fmt.Errorf("can not change your role")
		return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
	}

	if err := ah.srv.User.AddRole(cred); err != nil {
		log.Errorf("failed to add role: %v", err)
		wrapErr := fmt.Errorf("failed to add role: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return utils.Response(ctx, fiber.StatusOK, "OK")
}

// @Summary	profile
// @ID			profile
// @Accept		json
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	model.User		"Data"
// @Router		/profile [get]
func (ah *ApiHandler) profile(ctx *fiber.Ctx) error {
	userContext, err := ah.getUserFromContext(ctx)
	if err != nil {
		log.Errorf("failed to get user context: %v", err)
		wrapErr := fmt.Errorf("failed to get user context: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	user, err := ah.srv.User.GetUser(userContext.ID)
	if err != nil {
		log.Errorf("failed to get user: %v", err)
		wrapErr := fmt.Errorf("failed to get user: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return ctx.JSON(user)
}
