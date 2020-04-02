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
