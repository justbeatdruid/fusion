package driver

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	"k8s.io/klog"
)

func GetPostgresqlConnection(r *v1.RDB) (*sql.DB, error) {
	if r == nil {
		return nil, fmt.Errorf("datasource connect info is null")
	}
	//db, err := sql.Open("postgres", "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full")
	//db, err = sql.Open("postgres", "port=5433 user=postgres password=123456 dbname=ficow sslmode=disable")
	buildPath := strings.Builder{}
	buildPath.WriteString(fmt.Sprintf("host=%s ", r.Connect.Host))
	buildPath.WriteString(fmt.Sprintf("port=%d ", r.Connect.Port))
	buildPath.WriteString(fmt.Sprintf("user=%s ", r.Connect.Username))
	buildPath.WriteString(fmt.Sprintf("password=%s ", r.Connect.Password))
	buildPath.WriteString(fmt.Sprintf("dbname=%s ", r.Database))
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
