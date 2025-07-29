package model

type UploadBookCommand struct {
	Url string `json:"url"`
}

type BookInfo struct {
	Title      string
	Author     string
	Annotation string
}
