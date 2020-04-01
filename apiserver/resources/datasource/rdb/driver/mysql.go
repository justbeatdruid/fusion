package driver

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	"k8s.io/klog"
)

func GetMysqlData(ds *v1.Datasource, querySql string) ([]map[string]string, error) {
	db, err := GetMysqlConnection(ds)
	defer db.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to mysql: %+v", err)
	}
	data, err := GetMySQLDbData(db, querySql)
	if err != nil {
		return nil, fmt.Errorf("query by sql [%s] error: %+v", querySql, err)
	}
	return data, nil
}

func PingMysql(ds *v1.Datasource) error {
	db, err := GetMysqlConnection(ds)
	defer db.Close()
	if err != nil {
		return fmt.Errorf("cannot connect to mysql: %+v", err)
	}
	return PingMysqlConnection(db)
}

func GetMysqlConnection(ds *v1.Datasource) (*sql.DB, error) {
	if ds == nil || ds.Spec.RDB == nil {
		return nil, fmt.Errorf("datasource connect info is null")
	}
	buildPath := strings.Builder{}
	buildPath.WriteString(ds.Spec.RDB.Connect.Username)
	buildPath.WriteString(":")
	buildPath.WriteString(ds.Spec.RDB.Connect.Password)
	buildPath.WriteString("@tcp(")
	buildPath.WriteString(ds.Spec.RDB.Connect.Host)
	buildPath.WriteString(":")
	buildPath.WriteString(strconv.Itoa(ds.Spec.RDB.Connect.Port))
	buildPath.WriteString(")/")
	buildPath.WriteString(ds.Spec.RDB.Database)
	path := buildPath.String()
	klog.V(5).Infof("connection: %s", path)
	db, err := sql.Open("mysql", path)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %+v", err)
	}
	//设置数据库最大连接数
	db.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	db.SetMaxIdleConns(10)
	return db, nil
}

func PingMysqlConnection(db *sql.DB) error {
	//验证连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping database error: %+v", err)
	}
	return nil
}

func GetMySQLDbData(db *sql.DB, querySql string) ([]map[string]string, error) {
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
