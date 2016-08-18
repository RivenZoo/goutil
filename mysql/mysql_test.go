package mysql

import "testing"

var myConf = &MysqlConf{
	"test",
	"123456",
	"127.0.0.1:3306",
	"test",
	60,
}

func TestOpenMysql(t *testing.T) {
	db := NewMysqlDb(myConf)
	t.Log(myConf.dataSource())
	_, err := db.Exec("show tables;")
	t.Log(err)
}
