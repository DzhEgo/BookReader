package model

type Fb2TitleInfo struct {
	BookTitle string `xml:"book-title"`
	Author    struct {
		FirstName string `xml:"first-name"`
		LastName  string `xml:"last-name"`
	} `xml:"author"`
	Annotation struct {
		Paragraph string `xml:"p"`
	} `xml:"annotation"`
}

type Fb2Description struct {
	TitleInfo Fb2TitleInfo `xml:"title-info"`
}

type Fb2Root struct {
	Description Fb2Description `xml:"description"`
}
