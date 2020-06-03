package driver

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	"k8s.io/klog"
)

func GetRDBData(ds *v1.Datasource, querySql string) ([]map[string]string, error) {
	klog.V(5).Infof("query rdb with sql: %s", querySql)
	if ds == nil || ds.Spec.RDB == nil {
		return nil, fmt.Errorf("datasource connect info is null")
	}
	db, err := getConnection(ds, time.Duration(5*time.Second))
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
	db, err := getConnection(ds, time.Duration(5*time.Second))
	if err != nil || db == nil {
		return fmt.Errorf("cannot connect to database: %+v", err)
	}
	defer db.Close()
	return PingDatabase(db, time.Duration(5*time.Second))
}

func getConnection(ds *v1.Datasource, timeout time.Duration) (db *sql.DB, err error) {
	switch strings.ToLower(ds.Spec.RDB.Type) {
	case "mysql":
		db, err = GetMysqlConnection(ds)
	case "postgres", "postgresql":
		db, err = GetPostgresqlConnection(ds)
	default:
		err = fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
	}
	return
}

func PingDatabase(db *sql.DB, timeout time.Duration) (err error) {
	done := make(chan struct{})
	wait := time.After(timeout)
	go func() {
		klog.Infof("connect database in progress")
		//TODO conext cancel
		if err := db.Ping(); err != nil {
			err = fmt.Errorf("ping database error: %+v", err)
		}
		close(done)
	}()
	select {
	case <-wait:
		klog.Errorf("connection timeout, cancel")
		err = fmt.Errorf("connection timeout")
	case <-done:
		klog.Infof("connected")
	}
	return
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
