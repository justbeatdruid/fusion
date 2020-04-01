package driver

import (
	"fmt"
	"strings"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
)

func GetRDBData(ds *v1.Datasource, querySql string) ([]map[string]string, error) {
	if ds == nil || ds.Spec.RDB == nil {
		return nil, fmt.Errorf("datasource connect info is null")
	}
	switch strings.ToLower(ds.Spec.RDB.Type) {
	case "mysql":
		return GetMysqlData(ds, querySql)
	default:
		return nil, fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
	}
}

func PingRDB(ds *v1.Datasource) error {
	if ds == nil || ds.Spec.RDB == nil {
		return fmt.Errorf("datasource connect info is null")
	}
	switch strings.ToLower(ds.Spec.RDB.Type) {
	case "mysql":
		return PingMysql(ds)
	default:
		return fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
	}
}
