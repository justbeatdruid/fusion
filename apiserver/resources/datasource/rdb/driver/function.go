package driver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
)

func GetRDBData(ds *v1.Datasource, querySql string) ([]map[string]string, error) {
	if ds == nil || ds.Spec.RDB == nil {
		return nil, fmt.Errorf("datasource connect info is null")
	}
	var err error = nil
	var db *sql.DB
	switch strings.ToLower(ds.Spec.RDB.Type) {
	case "mysql":
		db, err = GetMysqlConnection(ds)
	case "postgres", "postgresql":
		db, err = GetPostgresqlConnection(ds)
	default:
		return nil, fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
	}
	if err != nil || db == nil {
		return nil, fmt.Errorf("cannot connect to database: %+v", err)
	}
	defer db.Close()
	return GetDatabaseData(db, querySql)
}

func PingRDB(ds *v1.Datasource) error {
	if ds == nil || ds.Spec.RDB == nil {
		return fmt.Errorf("datasource connect info is null")
	}
	var err error = nil
	var db *sql.DB
	switch strings.ToLower(ds.Spec.RDB.Type) {
	case "mysql":
		db, err = GetMysqlConnection(ds)
	case "postgres", "postgresql":
		db, err = GetPostgresqlConnection(ds)
	default:
		return fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
	}
	if err != nil || db == nil {
		return fmt.Errorf("cannot connect to database: %+v", err)
	}
	defer db.Close()
	return PingDatabase(db)
}

func PingDatabase(db *sql.DB) error {
	//验证连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping database error: %+v", err)
	}
	return nil
}

func GetDatabaseData(db *sql.DB, querySql string) ([]map[string]string, error) {
	rows, err := db.Query(querySql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//获取列名
	columns, _ := rows.Columns()

	//定义一个切片,长度是字段的个数,切片里面的元素类型是sql.RawBytes
	values := make([]sql.RawBytes, len(columns))
	//定义一个切片,元素类型是interface{} 接口
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		//把sql.RawBytes类型的地址存进去了
		scanArgs[i] = &values[i]
	}
	//获取字段值
	var result []map[string]string
	for rows.Next() {
		res := make(map[string]string)
		rows.Scan(scanArgs...)
		for i, col := range values {
			res[columns[i]] = string(col)
		}
		result = append(result, res)
	}

	return result, nil
}
