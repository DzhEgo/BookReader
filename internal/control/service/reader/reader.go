package reader

import (
	"BookStore/internal/control/model"
	"BookStore/internal/control/service/cache"
	"fmt"
	"math"
	"path/filepath"
	"strings"
)

const PageSize uint = 1500

type ReaderService struct {
	adapters map[string]BookReader
	cache    cache.MemoryCacheService
}

func NewService(cache cache.MemoryCacheService) *ReaderService {
	return &ReaderService{
		adapters: map[string]BookReader{
			"fb2":  &Fb2ReaderAdapter{},
			"epub": &EpubReaderAdapter{},
		},
		cache: cache,
	}
}

func (s *ReaderService) getAdapter(path string) (BookReader, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	adapter, ok := s.adapters[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	return adapter, nil
}

func (s *ReaderService) GetChaptersCount(path string) (uint, error) {
	key := fmt.Sprintf("bookChaptersCount:%s", path)
	if val, ok := s.cache.Get(key); ok {
		return val.(uint), nil
	}

	adapter, err := s.getAdapter(path)
	if err != nil {
		return 0, err
	}

	count, err := adapter.GetChaptersCount(path)
	if err != nil {
		return 0, err
	}

	s.cache.Set(key, count)

	return count, nil
}

func (s *ReaderService) Parse(path string) (string, error) {
	key := fmt.Sprintf("bookParse:%s", path)
	if val, ok := s.cache.Get(key); ok {
		return val.(string), nil
	}

	adapter, err := s.getAdapter(path)
	if err != nil {
		return "", err
	}

	data, err := adapter.Parse(path)
	if err != nil {
		return "", err
	}

	s.cache.Set(key, data)

	return data, nil
}

func (s *ReaderService) GetBookInfo(path string) (*model.BookInfo, error) {
	key := fmt.Sprintf("bookInfo:%s", path)
	if val, ok := s.cache.Get(key); ok {
		return val.(*model.BookInfo), nil
	}

	adapter, err := s.getAdapter(path)
	if err != nil {
		return nil, err
	}

	info, err := adapter.GetBookInfo(path)
	if err != nil {
		return nil, err
	}

	s.cache.Set(key, info)

	return info, nil
}

func (s *ReaderService) GetBookPage(path string, pageNum uint) (string, error) {
	key := fmt.Sprintf("%s:%d", path, pageNum)
	if val, ok := s.cache.Get(key); ok {
		return val.(string), nil
	}

	adapter, err := s.getAdapter(path)
	if err != nil {
		return "", err
	}

	data, err := s.Parse(path)
	if err != nil {
		return "", err
	}

	page, err := adapter.GetBookPage(data, pageNum)
	if err != nil {
		return "", err
	}

	s.cache.Set(key, page)

	return page, nil
}

func CountPages(runes []rune) uint {
	if len(runes) == 0 {
		return 0
	}

	return uint(math.Ceil(float64(len(runes)) / float64(PageSize)))
}
