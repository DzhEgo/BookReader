package model

import (
	"encoding/xml"
	"strings"
)

type Rootfile struct {
	FullPath string `xml:"full-path,attr"`
}

type Container struct {
	XMLName   xml.Name `xml:"container"`
	Rootfiles []struct {
		Rootfile Rootfile `xml:"rootfile"`
	} `xml:"rootfiles"`
}

type Package struct {
	Manifest struct {
		Items []Item `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		Itemrefs []Itemref `xml:"itemref"`
	} `xml:"spine"`
}

type Item struct {
	ID   string `xml:"id,attr"`
	Href string `xml:"href,attr"`
}

type Itemref struct {
	IDRef string `xml:"idref,attr"`
}

type TextExtractor struct {
	Tag     bool
	Data    []string
	Current strings.Builder
}
