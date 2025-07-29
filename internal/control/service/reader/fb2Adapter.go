package reader

import (
	"BookStore/internal/control/model"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

type Fb2ReaderAdapter struct{}

func (t *Fb2ReaderAdapter) Parse(path string) (string, error) {
	var content strings.Builder
	var inTitle, inParagraph, inSubtitle, inBody bool

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "body":
				inBody = true
			case "title":
				inTitle = true
			case "p":
				inParagraph = true
			case "subtitle":
				inSubtitle = true
			}
		case xml.EndElement:
			switch elem.Name.Local {
			case "title":
				inTitle = false
				content.WriteString("\n")
			case "p":
				inParagraph = false
				content.WriteString("\n\n")
			case "subtitle":
				inSubtitle = false
				content.WriteString("* * *\n\n")
			}
		case xml.CharData:
			if inBody {
				text := strings.TrimSpace(string(elem))
				if text != "" {
					switch {
					case inTitle:
						content.WriteString(text + "\n")
					case inParagraph:
						content.WriteString(text + " ")
					case inSubtitle:
					}
				}
			}
		}
	}

	return strings.TrimSpace(content.String()), nil
}

func (t *Fb2ReaderAdapter) GetChaptersCount(path string) (uint, error) {
	var count uint
	var inBody bool
	var subsection int

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Local == "body" {
				inBody = true
			}
			if inBody && elem.Name.Local == "section" {
				if subsection == 0 {
					count++
				}
				subsection++
			}
		case xml.EndElement:
			if elem.Name.Local == "body" {
				inBody = false
			}
			if inBody && elem.Name.Local == "section" {
				if subsection != 0 {
					subsection--
				}
			}
		}
	}

	return count, nil
}

func (t *Fb2ReaderAdapter) GetBookInfo(path string) (*model.BookInfo, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fb2 model.Fb2Root
	if err := xml.Unmarshal(data, &fb2); err != nil {
		return nil, err
	}

	info := fb2.Description.TitleInfo

	bookInfo := &model.BookInfo{
		Title:      info.BookTitle,
		Author:     info.Author.FirstName + " " + info.Author.LastName,
		Annotation: info.Annotation.Paragraph,
	}

	return bookInfo, nil
}

func (t *Fb2ReaderAdapter) GetBookPage(data string, pageNum uint) (string, error) {
	runes := []rune(data)
	totalPages := CountPages(runes)

	if pageNum < 1 || pageNum >= totalPages {
		return "", fmt.Errorf("page number out of range: %d", pageNum)
	}

	start := (pageNum - 1) * PageSize
	end := start + PageSize

	if end > uint(len(runes)) {
		end = uint(len(runes))
	}

	return string(runes[start:end]), nil
}
