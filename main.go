package main

import (
	"fmt"
	"log"
)

var (
	driver   = MYSQL
	username = "tkn_dev"
	password = "tkn_dev"
	url      = "localhost"
	port     = "3306"
	database = "tk_chat"

	mapperDir   = "./"
	mapperNames = MapperNames{"mapper.xml"}
)

type UserMst struct {
	UserId   string
	UserName string
}

func main() {

	dbConfig := NewDBConfig(driver, username, password, url, port, database)
	mapperConfig := NewMapperConfig(mapperDir, mapperNames)

	session, err := NewSQLSession(dbConfig, mapperConfig)

	if err != nil {
		log.Fatal(err)
	}

	entity := &UserMst{UserId: "song.of.restarting.2015@gmail.com"}

	result := &[]UserMst{}

	if err := session.SelectList(entity, result, "userMstMapper", "test"); err != nil {
		fmt.Println(err)
	}

	fmt.Println(result)
}
