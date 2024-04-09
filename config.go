package tksql

import (
	"path"
	"path/filepath"
	"regexp"
)

// MapperConfig マッパー設定構造体.
type MapperConfig struct {
	// mappersDir マッパーディレクトリ.
	mappersDir string
	// mapperNames マッパー名.
	mapperNames []string
}

// NewMapperConfig マッパー設定構造体を生成します.
func NewMapperConfig(mappersDir string, mapperNames []string) *MapperConfig {
	return &MapperConfig{
		mappersDir:  mappersDir,
		mapperNames: mapperNames,
	}
}

// NewMapperConfigGlob マッパー設定構造体を生成します(Globを使用しマッパーファイルを登録).
func NewMapperConfigGlob(mappersDir string, filePattern string) (*MapperConfig, error) {
	files, err := filepath.Glob(path.Join(mappersDir, filePattern))
	if err != nil {
		return nil, err
	}
	mappersNames := []string{}
	for _, file := range files {
		split := regexp.MustCompile(`(/|\\)`).Split(file, -1)
		mappersNames = append(mappersNames, split[len(split)-1])
	}
	return &MapperConfig{
		mappersDir:  mappersDir,
		mapperNames: mappersNames,
	}, nil
}

// DBConfig DB設定構造体.
type DBConfig struct {
	driver   Driver
	username string
	password string
	url      string
	port     string
	database string
	engine   string
	encoding string
}

// NewDBConfig DB設定構造体を生成します.
func NewDBConfig(driver Driver, username, password, url, port, database string) *DBConfig {
	return &DBConfig{
		driver:   driver,
		username: username,
		password: password,
		url:      url,
		port:     port,
		database: database,
	}
}

// SetEngine エンジンを設定します.
func (c *DBConfig) SetEngine(engine string) {
	c.engine = engine
}

// エンコーディングを設定します.
func (c *DBConfig) SetEnCoding(encoding string) {
	c.encoding = encoding
}
