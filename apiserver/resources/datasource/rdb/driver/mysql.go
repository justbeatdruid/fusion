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

func GetMysqlConnection(r *v1.RDB) (*sql.DB, error) {
	if r == nil {
		return nil, fmt.Errorf("datasource connect info is null")
	}
	buildPath := strings.Builder{}
	buildPath.WriteString(r.Connect.Username)
	buildPath.WriteString(":")
	buildPath.WriteString(r.Connect.Password)
	buildPath.WriteString("@tcp(")
	buildPath.WriteString(r.Connect.Host)
	buildPath.WriteString(":")
	buildPath.WriteString(strconv.Itoa(r.Connect.Port))
	buildPath.WriteString(")/")
	buildPath.WriteString(r.Database)
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
