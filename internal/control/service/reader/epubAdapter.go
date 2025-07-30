package reader

import (
	"BookStore/internal/control/model"
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"
)

type EpubReaderAdapter struct {
}

func (t *EpubReaderAdapter) Parse(path string) (string, error) {
	var content strings.Builder

	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	opfPath, err := t.getOpfPath(r)
	if err != nil {
		return "", err
	}

	manifest, spineIds, err := t.getManifestAndSpineIds(r, opfPath)
	if err != nil {
		return "", err
	}

	for _, id := range spineIds {
		href := manifest[id]
		if href != "cover.xhtml" && href != "" {
			for _, f := range r.File {
				if strings.HasSuffix(f.Name, href) {
					rc, err := f.Open()
					if err != nil {
						return "", err
					}
					defer rc.Close()

					data, err := io.ReadAll(rc)
					if err != nil {
						return "", err
					}

					text, err := t.textFromXhtml(string(data))
					if err != nil {
						continue
					}

					content.WriteString(text)
					content.WriteString("\n")
					break
				}
			}
		}
	}

	return content.String(), nil
}

func (t *EpubReaderAdapter) getOpfPath(r *zip.ReadCloser) (string, error) {
	var opfPath string
	for _, f := range r.File {
		if f.Name == "META-INF/container.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}

			var container model.Container
			if err := xml.Unmarshal(data, &container); err != nil {
				return "", err
			}

			opfPath = container.Rootfiles[0].Rootfile.FullPath
			break
		}
	}

	return opfPath, nil
}

func (t *EpubReaderAdapter) getManifestAndSpineIds(r *zip.ReadCloser, opfPath string) (map[string]string, []string, error) {
	var manifest = map[string]string{}
	var spineIds []string

	for _, f := range r.File {
		if f.Name == opfPath {
			rc, err := f.Open()
			if err != nil {
				return nil, nil, err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, nil, err
			}

			var pkg model.Package
			if err := xml.Unmarshal(data, &pkg); err != nil {
				return nil, nil, err
			}

			for _, item := range pkg.Manifest.Items {
				manifest[item.ID] = item.Href
			}

			for _, ref := range pkg.Spine.Itemrefs {
				spineIds = append(spineIds, ref.IDRef)
			}

			break
		}
	}

	return manifest, spineIds, nil
}

func (t *EpubReaderAdapter) GetChaptersCount(path string) (uint, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	var count uint
	for _, f := range r.File {
		if filepath.Ext(f.Name) == ".html" || filepath.Ext(f.Name) == ".xhtml" {
			if !strings.HasPrefix(f.Name, "cover") {
				count++
			}
		}
	}

	return count, nil
}

func (t *EpubReaderAdapter) GetBookInfo(path string) (*model.BookInfo, error) {
	var author, title, description string
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	opfPath, err := t.getOpfPath(r)
	if err != nil {
		return nil, err
	}

	if opfPath == "" {
		return nil, fmt.Errorf("opf path not found")
	}

	for _, f := range r.File {
		if f.Name == opfPath {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			dec := xml.NewDecoder(rc)
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}

				switch elem := tok.(type) {
				case xml.StartElement:
					switch elem.Name.Local {
					case "title":
						var val string
						if err := dec.DecodeElement(&val, &elem); err == nil && title == "" {
							title = strings.TrimSpace(val)
						}
					case "creator":
						var val string
						if err := dec.DecodeElement(&val, &elem); err == nil && author == "" {
							author = strings.TrimSpace(val)
						}
					case "description":
						var val string
						if err := dec.DecodeElement(&val, &elem); err == nil && description == "" {
							description = strings.TrimSpace(val)
						}
					}
				}
			}

			bookInfo := &model.BookInfo{
				Title:      title,
				Author:     author,
				Annotation: description,
			}

			return bookInfo, nil
		}
	}
	return nil, fmt.Errorf("failed to get book info")
}

func (t *EpubReaderAdapter) GetBookPage(data string, pageNum uint) (string, error) {
	runes := []rune(data)
	length := uint(len(runes))
	var start uint
	var end uint

	if length == 0 {
		return "", nil
	}

	for i := uint(1); i < pageNum; i++ {
		tmpEnd := start + PageSize
		if tmpEnd >= length {
			tmpEnd = length
		} else {
			for tmpEnd < length && !unicode.IsSpace(runes[tmpEnd]) {
				tmpEnd++
			}
		}
		start = tmpEnd
	}

	end = start + PageSize
	if end >= length {
		end = length
	} else {
		for end < length && !unicode.IsSpace(runes[end]) {
			end++
		}
	}

	if start >= length {
		return "", fmt.Errorf("start out of bounds")
	}

	return strings.TrimSpace(string(runes[start:end])), nil
}

func (t *EpubReaderAdapter) textFromXhtml(data string) (string, error) {
	var extractor model.TextExtractor
	dec := xml.NewDecoder(strings.NewReader(data))

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Local == "p" || elem.Name.Local == "div" {
				extractor.Tag = true
				extractor.Current.Reset()
			}
		case xml.CharData:
			if extractor.Tag {
				extractor.Current.Write(elem)
			}
		case xml.EndElement:
			if extractor.Tag && (elem.Name.Local == "p" || elem.Name.Local == "div") {
				extractor.Tag = false
				raw := strings.TrimSpace(extractor.Current.String())
				if raw != "" {
					extractor.Data = append(extractor.Data, raw)
				}
			}
		}
	}

	return strings.Join(extractor.Data, "\n\n"), nil
}
