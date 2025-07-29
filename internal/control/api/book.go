package api

import (
	"BookStore/internal/common/utils"
	"BookStore/internal/control/model"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"strconv"
	"strings"
)

// @Summary	upload book
// @ID			uploadBook
// @Accept		json
// @Param		params	body		model.UploadBookCommand	true	"User credentials"	request
// @Failure	500		{object}	model.Response			"Internal Server Error"
// @Failure	400		{object}	model.Response			"Bad Request"
// @Failure	401		{object}	model.Response			"Unauthorized"
// @Success	200		{object}	string					"OK"
// @Router		/book/upload [post]
func (ah *ApiHandler) uploadBook(ctx *fiber.Ctx) error {
	user, err := ah.getUserFromContext(ctx)
	if err != nil {
		log.Errorf("failed to get user: %v", err)
		wrapErr := fmt.Errorf("failed to get user: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	if user.Role == "user" {
		log.Errorf("forbidden to upload book")
		wrapErr := fmt.Errorf("forbidden to upload book")
		return utils.Response(ctx, fiber.StatusForbidden, wrapErr.Error())
	}

	contentType := ctx.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		file, err := ctx.FormFile("file")
		if err != nil {
			log.Errorf("failed to get file: %v", err)
			wrapErr := fmt.Errorf("failed to get file: %v", err)
			return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
		}

		if err := ah.srv.Books.UploadBookLocal(ctx, file, user); err != nil {
			log.Errorf("failed to upload book: %v", err)
			wrapErr := fmt.Errorf("failed to upload book: %v", err)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}

		return utils.Response(ctx, fiber.StatusOK, "OK")
	}

	if strings.Contains(contentType, "application/json") {
		var url model.UploadBookCommand

		if err := json.Unmarshal(ctx.Body(), &url); err != nil {
			log.Errorf("failed to unmarshal json: %v", err)
			wrapErr := fmt.Errorf("failed to unmarshal json: %v", err)
			return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
		}

		if err := ah.srv.Books.UploadBookUrl(url, user); err != nil {
			log.Errorf("failed to upload book: %v", err)
			wrapErr := fmt.Errorf("failed to upload book: %v", err)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}

		return utils.Response(ctx, fiber.StatusOK, "OK")
	}

	return utils.Response(ctx, fiber.StatusBadRequest, "invalid content-type")
}

// @Summary	get books
// @ID			getBook
// @Accept		json
// @Param		id	query		int				true	"Book id"	request
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	model.Book		"Data"
// @Router		/book/list [get]
func (ah *ApiHandler) getBook(ctx *fiber.Ctx) error {
	id := ctx.Query("id")

	if id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Errorf("failed to convert id to int: %v", err)
			wrapErr := fmt.Errorf("failed to get book id: %v", err)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}

		book, err := ah.srv.Books.GetBook(idInt)
		if err != nil {
			log.Errorf("failed to get book: %v", err)
			wrapErr := fmt.Errorf("failed to get book: %v", err)
			return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
		}

		return ctx.JSON(book)
	}

	books, err := ah.srv.Books.GetBooks()
	if err != nil {
		log.Errorf("failed to get books: %v", err)
		wrapErr := fmt.Errorf("failed to get books: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return ctx.JSON(books)
}

// @Summary	delete book
// @ID			deleteBook
// @Accept		json
// @Param		id	query		int				true	"Book id"	request
// @Failure	500	{object}	model.Response	"Internal Server Error"
// @Failure	400	{object}	model.Response	"Bad Request"
// @Failure	401	{object}	model.Response	"Unauthorized"
// @Success	200	{object}	string			"OK"
// @Router		/book/delete [delete]
func (ah *ApiHandler) deleteBook(ctx *fiber.Ctx) error {
	id := ctx.Query("id")

	if id == "" {
		log.Errorf("failed to get book id")
		wrapErr := fmt.Errorf("failed to get book id")
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Errorf("failed to convert id to int: %v", err)
		wrapErr := fmt.Errorf("failed to convert id to int: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	user, err := ah.getUserFromContext(ctx)
	if err != nil {
		log.Errorf("failed to get user: %v", err)
		wrapErr := fmt.Errorf("failed to get user: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	if err := ah.srv.Books.DeleteBook(idInt, user); err != nil {
		log.Errorf("failed to delete book: %v", err)
		wrapErr := fmt.Errorf("failed to delete book: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return utils.Response(ctx, fiber.StatusOK, "OK")
}

// @Summary	get book page
// @ID			getBookPage
// @Accept		json
// @Param		id		query		int				true	"Book id"		request
// @Param		page	query		int				true	"Page number"	request
// @Failure	500		{object}	model.Response	"Internal Server Error"
// @Failure	400		{object}	model.Response	"Bad Request"
// @Failure	401		{object}	model.Response	"Unauthorized"
// @Success	200		{object}	string			"OK"
// @Router		/book/read [get]
func (ah *ApiHandler) getBookPage(ctx *fiber.Ctx) error {
	id := ctx.Query("id")
	page := ctx.Query("page")
	if id == "" || page == "" {
		log.Errorf("failed to get book page")
		wrapErr := fmt.Errorf("failed to get book page")
		return utils.Response(ctx, fiber.StatusBadRequest, wrapErr.Error())
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Errorf("failed to convert id to int: %v", err)
		wrapErr := fmt.Errorf("failed to convert id to int: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		log.Errorf("failed to convert page to int: %v", err)
		wrapErr := fmt.Errorf("failed to convert page to int: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	user, err := ah.getUserFromContext(ctx)
	if err != nil {
		log.Errorf("failed to get user: %v", err)
		wrapErr := fmt.Errorf("failed to get user: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	if user == nil {
		log.Errorf("unauthorized")
		wrapErr := fmt.Errorf("unauthorized")
		return utils.Response(ctx, fiber.StatusUnauthorized, wrapErr.Error())
	}

	if (user.Role != "super" && user.Role != "admin") && pageInt > 15 {
		log.Errorf("not allowed")
		wrapErr := fmt.Errorf("not allowed")
		return utils.Response(ctx, fiber.StatusForbidden, wrapErr.Error())
	}

	book, err := ah.srv.Books.GetBook(idInt)
	if err != nil {
		log.Errorf("failed to get book: %v", err)
		wrapErr := fmt.Errorf("failed to get book: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	bookPage, err := ah.srv.Reader.GetBookPage(book.Filepath, uint(pageInt))
	if err != nil {
		log.Errorf("failed to get book page: %v", err)
		wrapErr := fmt.Errorf("failed to get book page: %v", err)
		return utils.Response(ctx, fiber.StatusInternalServerError, wrapErr.Error())
	}

	return ctx.JSON(bookPage)
}
