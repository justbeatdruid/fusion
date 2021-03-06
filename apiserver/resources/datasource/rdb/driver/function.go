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
	return GetRDBDatabaseData(ds.Spec.RDB, querySql)
}

func GetRDBDatabaseData(rdb *v1.RDB, querySql string) ([]map[string]string, error) {
	if rdb == nil {
		return nil, fmt.Errorf("rdb is null")
	}
	db, err := getConnection(rdb, time.Duration(5*time.Second))
	if err != nil || db == nil {
		return nil, fmt.Errorf("cannot connect to database: %+v", err)
	}
	defer db.Close()
	return GetDatabaseData(db, querySql)
}

func PingRDB(ds *v1.Datasource) error {
	if ds == nil {
		return fmt.Errorf("datasource is null")
	}
	return PingRDBDatabase(ds.Spec.RDB)
}

func PingRDBDatabase(r *v1.RDB) error {
	if r == nil {
		return fmt.Errorf("datasource connect info is null")
	}
	db, err := getConnection(r, time.Duration(5*time.Second))
	if err != nil || db == nil {
		return fmt.Errorf("cannot connect to database: %+v", err)
	}
	defer db.Close()
	return PingDatabase(db, time.Duration(5*time.Second))
}

func getConnection(r *v1.RDB, timeout time.Duration) (db *sql.DB, err error) {
	switch strings.ToLower(r.Type) {
	case "mysql":
		db, err = GetMysqlConnection(r)
	case "postgres", "postgresql":
		db, err = GetPostgresqlConnection(r)
	default:
		err = fmt.Errorf("unsupported rdb type %s", r.Type)
	}
	return
}

func PingDatabase(db *sql.DB, timeout time.Duration) (err error) {
	done := make(chan struct{})
	wait := time.After(timeout)
	go func() {
		klog.V(5).Infof("connect database in progress")
		//TODO conext cancel
		if e := db.Ping(); e != nil {
			err = fmt.Errorf("ping database error: %+v", e)
		}
		close(done)
	}()
	select {
	case <-wait:
		klog.Errorf("connection timeout, cancel")
		err = fmt.Errorf("connection timeout")
	case <-done:
		if err != nil {
			klog.V(5).Infof("%+v", err)
		} else {
			klog.V(5).Infof("connected")
		}
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
