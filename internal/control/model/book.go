package model

type UploadBookCommand struct {
	Url string `json:"url"`
}

type BookInfo struct {
	Title      string
	Author     string
	Annotation string
}

type SaveProgress struct {
	BookId int `json:"book_id"`
	Page   int `json:"page"`

	userId int
}

func (s *SaveProgress) SetUserId(userId int) {
	s.userId = userId
}

func (s *SaveProgress) GetUserId() int {
	return s.userId
}
