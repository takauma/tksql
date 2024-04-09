package tksql

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// mapper マッパー構造体.
type mapper struct {
	Name   string  `xml:"name,attr"`
	Select []query `xml:"select"`
	Insert []query `xml:"insert"`
	Update []query `xml:"update"`
	Delete []query `xml:"delete"`
}

// query クエリ構造体.
type query struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",cdata"`
}

// parseMapper XMLファイルを解析しマッパー構造体に格納します.
func parseMapper(mapperDir, filename string) (*mapper, error) {
	// XMLファイルの内容を読み込む.
	file, err := os.Open(path.Join(mapperDir, filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fmt.Println("[tksql] Loaded mapper file: " + strings.ReplaceAll(file.Name(), `\`, "/"))

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// マッパー構造体.
	mapper := &mapper{}

	// XMLファイルを解析して構造体に格納.
	if err = xml.Unmarshal(data, mapper); err != nil {
		return nil, err
	}

	for _, query := range mapper.Insert {
		fmt.Println("[tksql] Loaded query: " + mapper.Name + "." + query.ID)
	}
	for _, query := range mapper.Update {
		fmt.Println("[tksql] Loaded query: " + mapper.Name + "." + query.ID)
	}
	for _, query := range mapper.Delete {
		fmt.Println("[tksql] Loaded query: " + mapper.Name + "." + query.ID)
	}
	for _, query := range mapper.Select {
		fmt.Println("[tksql] Loaded query: " + mapper.Name + "." + query.ID)
	}

	return mapper, nil
}
