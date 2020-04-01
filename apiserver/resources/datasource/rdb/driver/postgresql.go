package driver

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	"k8s.io/klog"
)

func GetPostgresqlConnection(ds *v1.Datasource) (*sql.DB, error) {
	if ds == nil || ds.Spec.RDB == nil {
		return nil, fmt.Errorf("datasource connect info is null")
	}
	//db, err := sql.Open("postgres", "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full")
	//db, err = sql.Open("postgres", "port=5433 user=postgres password=123456 dbname=ficow sslmode=disable")
	buildPath := strings.Builder{}
	buildPath.WriteString(fmt.Sprintf("host=%s ", ds.Spec.RDB.Connect.Host))
	buildPath.WriteString(fmt.Sprintf("port=%d ", ds.Spec.RDB.Connect.Port))
	buildPath.WriteString(fmt.Sprintf("user=%s ", ds.Spec.RDB.Connect.Username))
	buildPath.WriteString(fmt.Sprintf("password=%s ", ds.Spec.RDB.Connect.Password))
	buildPath.WriteString(fmt.Sprintf("dbname=%s ", ds.Spec.RDB.Database))
	buildPath.WriteString(fmt.Sprintf("sslmode=%s ", "disable"))
	path := buildPath.String()
	klog.V(5).Infof("connection: %s", path)
	db, err := sql.Open("postgres", path)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %+v", err)
	}
	//设置数据库最大连接数
	db.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	db.SetMaxIdleConns(10)
	return db, nil
}
