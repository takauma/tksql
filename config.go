package main

// MapperNames マッパー名リスト.
type MapperNames []string

// MapperConfig マッパー設定構造体.
type MapperConfig struct {
	mappersDir  string
	mapperNames MapperNames
}

// NewMapperConfig マッパー設定構造体を生成します.
func NewMapperConfig(mappersDir string, mapperNames MapperNames) *MapperConfig {
	return &MapperConfig{
		mappersDir:  mappersDir,
		mapperNames: mapperNames,
	}
}

// DBConfig DB設定構造体.
type DBConfig struct {
	driver   Driver
	username string
	password string
	url      string
	port     string
	database string
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
