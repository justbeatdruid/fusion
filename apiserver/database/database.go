package database

import (
	"fmt"
	"sync"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"k8s.io/klog"
)

type DatabaseConnection struct {
	orm.Ormer
	mtx *sync.Mutex
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
	return &DatabaseConnection{orm.NewOrm(), &sync.Mutex{}}, nil
}

func (d *DatabaseConnection) Begin() error {
	d.mtx.Lock()
	if err := d.Ormer.Begin(); err != nil {
		d.mtx.Unlock()
		return err
	}
	return nil
}

func (d *DatabaseConnection) Commit() error {
	defer d.mtx.Unlock()
	return d.Ormer.Commit()
}

func (d *DatabaseConnection) AddObject(o model.Table, us interface{}, count int) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Insert(o)
	if err != nil {
		d.Rollback()
		return err
	}
	_, err = d.InsertMulti(count, us)
	if err != nil {
		d.Rollback()
		return err
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) UpdateObject(o model.Table, us interface{}, count int) (err error) {
	if err := d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	if _, err := d.Update(o); err != nil {
		d.Rollback()
		return err
	}
	if us != nil {
		ul := []model.UserRelation{}
		if _, err := d.QueryTable("UserRelation").Filter("ResourceId", o.ResourceType()).Filter("ResourceType", o.ResourceType()).All(&ul); err != nil {
			d.Rollback()
			return err
		}
		for _, u := range ul {
			if _, err := d.Delete(&u); err != nil {
				d.Rollback()
				return err
			}
		}
		_, err = d.InsertMulti(count, us)
		if err != nil {
			d.Rollback()
			return err
		}
	}
	if err != nil {
		d.Rollback()
		return err
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) DeleteObject(o model.Table) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Delete(o)
	ul := []model.UserRelation{}
	if _, err := d.QueryTable("UserRelation").Filter("ResourceId", o.ResourceId()).Filter("ResourceType", o.ResourceType()).All(&ul); err != nil {
		d.Rollback()
		return err
	}
	for _, u := range ul {
		if _, err := d.Delete(&u); err != nil {
			d.Rollback()
			return err
		}
	}
	if err != nil {
		d.Rollback()
		return err
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) AddApplication(obj interface{}) error {
	o, us, err := model.ApplicationGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get application from obj error: %+v", err)
	}
	return d.AddObject(&o, us, len(us))
}

func (d *DatabaseConnection) UpdateApplication(old, obj interface{}) error {
	_, ous, err := model.ApplicationGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get users from obj error: %+v", err)
	}
	o, us, err := model.ApplicationGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get application from obj error: %+v", err)
	}
	if model.Equal(ous, us) {
		return d.UpdateObject(&o, nil, 0)
	}

	return d.UpdateObject(&o, us, len(us))
}

func (d *DatabaseConnection) DeleteApplication(obj interface{}) error {
	o, _, err := model.ApplicationGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get application from obj error: %+v", err)
	}
	return d.DeleteObject(&o)
}
