package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const (
	sqlPrepareStr = `?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,`
	baseInsertSql = `insert into %s (%s) values (%s)`
	baseSelectSql = `select %s from %s %s`
	baseDeleteSql = `delete from %s where %s`
	baseUpdateSql = `update %s set %s=? where %s`
)

var (
	sqlPrepareStrLen = len(sqlPrepareStr)
	maxFields        = sqlPrepareStrLen / 2
)

var (
	SqlErrNeedTableName  = errors.New("need table name")
	SqlErrNeedValues     = errors.New("need values")
	SqlErrNeedConditions = errors.New("need conditions")
	SqlErrOverMax        = errors.New("over max field:32")
	SqlErrNeedFields     = errors.New("need fields")
)

// max n=32
func prepareStrLen(n int) int {
	if n <= 0 && n > maxFields {
		return -1
	}
	return 2*n - 1
}

func buildInsertSql(table *string, fields []string, valuesLen int) (*string, error) {
	if table == nil || *table == "" {
		return nil, SqlErrNeedTableName
	}

	if valuesLen > maxFields {
		return nil, SqlErrOverMax
	}

	pLen := prepareStrLen(valuesLen)
	sqlStr := fmt.Sprintf(baseInsertSql, *table, strings.Join(fields, ","), sqlPrepareStr[:pLen])
	return &sqlStr, nil
}

// max field number:32
func Insert(db *sql.DB, table *string, fields []string, values ...interface{}) (sql.Result, error) {
	sqlStr, err := buildInsertSql(table, fields, len(values))
	if err != nil {
		return nil, err
	}
	return db.Exec(*sqlStr, values...)
}

// if len(fields)==0, equal select *
// conditions: [field1='?',...]
// condition:'where ' +join(conditions,' and ')
// if no condition, will append 'limit 1000'
func buildSelectSql(table *string, fields []string, conditions []string) (*string, error) {
	if table == nil || *table == "" {
		return nil, SqlErrNeedTableName
	}
	var part1 string
	if len(fields) == 0 {
		part1 = "*"
	} else {
		part1 = strings.Join(fields, ",")
	}
	var part3 string
	if len(conditions) == 0 {
		part3 = "limit 1000"
	} else {
		part3 = "where " + strings.Join(conditions, " and ")
	}

	sqlStr := fmt.Sprintf(baseSelectSql, part1, *table, part3)
	return &sqlStr, nil
}

// should: len(values)==len(conditions)
func Select(db *sql.DB, table *string, fields []string, conditions []string, values ...interface{}) (*sql.Rows, error) {
	sqlStr, err := buildSelectSql(table, fields, conditions)
	if err != nil {
		return nil, err
	}
	return db.Query(*sqlStr, values...)
}

// should: len(values)==len(conditions)
// just return one row
func SelectRow(db *sql.DB, table *string, fields []string, conditions []string, values ...interface{}) *sql.Row {
	sqlStr, err := buildSelectSql(table, fields, conditions)
	if err != nil {
		return nil
	}
	return db.QueryRow(*sqlStr, values...)
}

// if len(conditions)==0, return error
// conditions: [field1='?',...]
// condition: join(conditions,' and ')
func buildDeleteSql(table *string, conditions []string) (*string, error) {
	if table == nil || *table == "" {
		return nil, SqlErrNeedTableName
	}
	if len(conditions) == 0 {
		return nil, SqlErrNeedConditions
	}
	part2 := strings.Join(conditions, " and ")
	sqlStr := fmt.Sprintf(baseDeleteSql, *table, part2)
	return &sqlStr, nil
}

// should: len(values)==len(conditions)
func Delete(db *sql.DB, table *string, conditions []string, values ...interface{}) (sql.Result, error) {
	sqlStr, err := buildDeleteSql(table, conditions)
	if err != nil {
		return nil, err
	}
	return db.Exec(*sqlStr, values...)
}

// if len(fields)==0, return error
// if len(conditions)==0, return error
// conditions: [field1='?',...]
// condition: join(conditions,' and ')
func buildUpdateSql(table *string, fields []string, conditions []string) (*string, error) {
	if table == nil || *table == "" {
		return nil, SqlErrNeedTableName
	}
	if len(fields) == 0 {
		return nil, SqlErrNeedFields
	}
	if len(conditions) == 0 {
		return nil, SqlErrNeedConditions
	}
	part2 := strings.Join(fields, "=?,")
	part3 := strings.Join(conditions, " and ")
	sqlStr := fmt.Sprintf(baseUpdateSql, *table, part2, part3)
	return &sqlStr, nil
}

// args: field value and condition value
func Update(db *sql.DB, table *string, fields []string, conditions []string, args ...interface{}) (sql.Result, error) {
	sqlStr, err := buildUpdateSql(table, fields, conditions)
	if err != nil {
		return nil, err
	}
	return db.Exec(*sqlStr, args...)
}
