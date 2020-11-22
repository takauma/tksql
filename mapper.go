package main

import (
	"encoding/xml"
	"io/ioutil"
)

// Mapper Mapper構造体.
type mapper struct {
	Name   string    `xml:"name,attr"`
	Select []sSelect `xml:"select"`
	Insert []sInsert `xml:"insert"`
	Update []sUpdate `xml:"update"`
	Delete []sDelete `xml:"delete"`
}

// sSelect Select構造体.
type sSelect struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr"`
	ResultType    string `xml:"resultType,attr"`
	Value         string `xml:",cdata"`
}

// sInsert Insert構造体.
type sInsert struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr"`
	Value         string `xml:",cdata"`
}

// sUpdate Update構造体.
type sUpdate struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr"`
	Value         string `xml:",cdata"`
}

// sDelete Delete構造体.
type sDelete struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr"`
	Value         string `xml:",cdata"`
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
