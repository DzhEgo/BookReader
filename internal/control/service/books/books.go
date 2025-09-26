package books

import (
	"BookStore/internal/control/model"
	"BookStore/internal/control/service/cache"
	"BookStore/internal/control/service/reader"
	dbmodel "BookStore/internal/database/model"
	"BookStore/internal/database/table"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BookService interface {
	UploadBookLocal(ctx *fiber.Ctx, file *multipart.FileHeader, user *model.UserContext) error
	UploadBookUrl(book model.UploadBookCommand, user *model.UserContext) error
	GetBook(id int) (*dbmodel.Book, error)
	GetBooks() ([]*dbmodel.Book, error)
	DeleteBook(id int, user *model.UserContext) error
	SaveProgress(command *model.SaveProgress) error
	GetProgress(userId, bookId int) (*dbmodel.ReadingProgress, error)
}
type Option func(*bookService)

type bookService struct {
	reader reader.BookReader
	cache  cache.MemoryCacheService
}

func NewService(opts ...Option) BookService {
	s := bookService{}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

func WithReader(r reader.BookReader) Option {
	return func(s *bookService) {
		s.reader = r
	}
}

func WithCache(c cache.MemoryCacheService) Option {
	return func(s *bookService) {
		s.cache = c
	}
}

func (b *bookService) UploadBookLocal(ctx *fiber.Ctx, file *multipart.FileHeader, user *model.UserContext) error {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == ".fb2" || ext == ".epub" {
		createTime := time.Now().Unix()
		dir := fmt.Sprintf("/var/tmp/%s", user.Login)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir: %v", err)
		}

		dest := filepath.Join(dir, fmt.Sprintf("%d_%s", createTime, file.Filename))
		if err := ctx.SaveFile(file, dest); err != nil {
			return fmt.Errorf("failed to upload file: %v", err)
		}

		count, err := b.reader.GetChaptersCount(dest)
		if err != nil {
			return fmt.Errorf("failed to get chapter count: %v", err)
		}

		bookInfo, err := b.reader.GetBookInfo(dest)
		if err != nil {
			return fmt.Errorf("failed to get book info: %v", err)
		}

		text, err := b.reader.Parse(dest)
		if err != nil {
			return fmt.Errorf("failed to parse book: %v", err)
		}

		runes := []rune(text)
		pages := reader.CountPages(runes)

		bookDb := &dbmodel.Book{
			Title:      bookInfo.Title,
			Format:     ext,
			Author:     bookInfo.Author,
			Annotation: bookInfo.Annotation,
			Filepath:   dest,
			Chapters:   count,
			Pages:      pages,
			CreatedAt:  createTime,
			UserId:     user.ID,
		}

		if err := table.Upsert(bookDb); err != nil {
			return fmt.Errorf("failed to upsert book: %v", err)
		}

		b.cache.Delete("allBooks")

		return nil
	}

	return fmt.Errorf("unsupported file extension")
}

func (b *bookService) UploadBookUrl(book model.UploadBookCommand, user *model.UserContext) error {
	if book.Url == "" {
		return fmt.Errorf("empty url")
	}

	resp, err := http.Get(book.Url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload book: %v", err)
	}
	defer resp.Body.Close()

	sp := strings.Split(book.Url, "/")
	fileName := sp[len(sp)-1]

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == ".fb2" || ext == ".epub" {
		createTime := time.Now().Unix()
		dir := fmt.Sprintf("/var/tmp/%s", user.Login)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir: %v", err)
		}

		dest := filepath.Join(dir, fmt.Sprintf("%d_%s", createTime, fileName))
		out, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("failed to upload file: %v", err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to upload file: %v", err)
		}

		count, err := b.reader.GetChaptersCount(dest)
		if err != nil {
			return fmt.Errorf("failed to get chapter count: %v", err)
		}

		bookInfo, err := b.reader.GetBookInfo(dest)
		if err != nil {
			return fmt.Errorf("failed to get book info: %v", err)
		}

		text, err := b.reader.Parse(dest)
		if err != nil {
			return fmt.Errorf("failed to parse book: %v", err)
		}

		runes := []rune(text)
		pages := reader.CountPages(runes)

		bookDb := &dbmodel.Book{
			Title:      bookInfo.Title,
			Format:     ext,
			Author:     bookInfo.Author,
			Annotation: bookInfo.Annotation,
			Filepath:   dest,
			Chapters:   count,
			Pages:      pages,
			CreatedAt:  createTime,
			UserId:     user.ID,
		}

		if err := table.Upsert(bookDb); err != nil {
			return fmt.Errorf("failed to upsert book: %v", err)
		}

		b.cache.Delete("allBooks")

		return nil
	}

	return fmt.Errorf("invalid file extension")
}

func (b *bookService) GetBook(id int) (*dbmodel.Book, error) {
	key := fmt.Sprintf("bookId:%d", id)
	if val, ok := b.cache.Get(key); ok {
		return val.(*dbmodel.Book), nil
	}

	book, err := table.GetBook(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get book: %v", err)
	}

	b.cache.Set(key, book)

	return book, nil
}

func (b *bookService) GetBooks() ([]*dbmodel.Book, error) {
	key := "allBooks"
	if val, ok := b.cache.Get(key); ok {
		return val.([]*dbmodel.Book), nil
	}

	books, err := table.GetBooks()
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %v", err)
	}

	b.cache.Set(key, books)

	return books, nil
}

func (b *bookService) DeleteBook(id int, user *model.UserContext) (err error) {
	var book *dbmodel.Book
	key := fmt.Sprintf("bookId:%d", id)

	val, ok := b.cache.Get(key)
	if !ok {
		book, err = table.GetBook(id)
		if err != nil {
			return fmt.Errorf("failed to get book: %v", err)
		}
	} else {
		book = val.(*dbmodel.Book)
	}

	if user.Role != "admin" {
		if user.ID != book.UserId {
			return fmt.Errorf("you have no permission to delete book")
		}
	}

	if err := table.DeleteBook(id); err != nil {
		return fmt.Errorf("failed to delete book: %v", err)
	}

	if err := os.Remove(book.Filepath); err != nil {
		return fmt.Errorf("failed to delete book: %v", err)
	}

	b.cache.Delete(key)
	b.cache.Delete("allBooks")

	return nil
}

func (b *bookService) SaveProgress(command *model.SaveProgress) error {
	existProgress, err := table.GetProgress(command.GetUserId(), command.BookId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		progress := &dbmodel.ReadingProgress{
			UserID:      command.GetUserId(),
			BookID:      command.BookId,
			CurrentPage: command.Page,
			LastReadAt:  time.Now().Unix(),
		}

		return table.Upsert(progress)
	}

	if existProgress != nil {
		return table.UpdateProgress(existProgress, command.Page)
	}

	return nil
}

func (b *bookService) GetProgress(userId, bookId int) (*dbmodel.ReadingProgress, error) {
	return table.GetProgress(userId, bookId)
}
