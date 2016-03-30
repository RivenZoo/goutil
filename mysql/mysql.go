package model

import "database/sql"
import (
	_mysql "github.com/go-sql-driver/mysql"
	"fmt"
)

type Logger interface {
	Print(v ...interface{})
}

func SetMysqlLogger(logger Logger) error {
	return _mysql.SetLogger(_mysql.Logger(logger))
}

type MysqlConf struct {
	User, Password, Addr, DbName string
	Timeout int // seconds
}

func (c MysqlConf) dataSource() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%ds&charset=utf8",
		c.User, c.Password, c.Addr, c.DbName, c.Timeout)
}

func NewMysqlDb(conf *MysqlConf) *sql.DB {
	db, err := sql.Open("mysql", conf.dataSource())
	if err != nil {
		panic(err)
	}
	return db
}

func CloseDb(db *sql.DB) {
	db.Close()
}