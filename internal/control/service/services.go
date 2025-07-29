package service

import (
	"BookStore/internal/control/service/auth"
	"BookStore/internal/control/service/books"
	"BookStore/internal/control/service/cache"
	"BookStore/internal/control/service/reader"
	"BookStore/internal/control/service/users"
)

type Services struct {
	Books  books.BookService
	Auth   auth.AuthService
	User   users.UserService
	Reader reader.BookReader
	Cache  cache.MemoryCacheService
}
