package reader

import "BookStore/internal/control/model"

type BookReader interface {
	Parse(path string) (string, error)
	GetChaptersCount(path string) (uint, error)
	GetBookInfo(path string) (*model.BookInfo, error)
	GetBookPage(data string, pageNum uint) (string, error)
}
