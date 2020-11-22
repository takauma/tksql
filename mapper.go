package tksql

import (
	"encoding/xml"
	"io/ioutil"
)

// Mapper Mapper構造体.
type mapper struct {
	Name   string  `xml:"name,attr"`
	Select []query `xml:"select"`
	Insert []query `xml:"insert"`
	Update []query `xml:"update"`
	Delete []query `xml:"delete"`
}

// sSelect Select構造体.
type query struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",cdata"`
}

// parseMapper XMLファイルを解析しマッパー構造体に格納します.
func parseMapper(path, filename string) (*mapper, error) {
	// XMLファイルの内容を読み込む.
	data, err := ioutil.ReadFile(path + filename)

	if err != nil {
		return nil, err
	}

	// マッパー構造体.
	mapper := &mapper{}

	// XMLファイルを解析して構造体に格納.
	if err = xml.Unmarshal(data, mapper); err != nil {
		return nil, err
	}

	return mapper, nil
}
