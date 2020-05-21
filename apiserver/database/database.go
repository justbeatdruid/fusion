package database

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"k8s.io/klog"
)

type DatabaseConnection struct {
}

// implements io.Writer
type KlogWriter struct {
}

func (*KlogWriter) Write(p []byte) (int, error) {
	klog.V(5).Infof(string(p))
	return len(p), nil
}

const databaseAlias = "default"

func NewDatabaseConnection(cfg *config.DatabaseConfig) (*DatabaseConnection, error) {
	orm.DebugLog = orm.NewLog(new(KlogWriter))
	orm.Debug = true
	var err error
	switch cfg.Type {
	case "mysql":
		orm.RegisterDriver("mysql", orm.DRMySQL)
		err = orm.RegisterDataBase(databaseAlias, "mysql",
			fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", cfg.Username, cfg.Password,
				cfg.Host, cfg.Port, cfg.Database))
	case "postgres":
		orm.RegisterDriver("postgres", orm.DRPostgres)
		err = orm.RegisterDataBase(databaseAlias, "postgres",
			fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
				cfg.Host, cfg.Port,
				cfg.Username, cfg.Password, cfg.Database))
	default:
		return nil, fmt.Errorf("unspported database %s", cfg.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot register database: %+v", err)
	}
	orm.RegisterModel(new(model.Application), new(model.UserRelation))
	if err = orm.RunSyncdb("default", false, true); err != nil {
		return nil, fmt.Errorf("cannot sync database: %+v", err)
	}
	return &DatabaseConnection{}, nil
}
