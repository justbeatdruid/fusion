package database

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/chinamobile/nlpt/apiserver/database/model"

	"k8s.io/klog"
)

type DatabaseConnection struct {
	orm.Ormer
	mtx *sync.Mutex

	enabled bool
}

// implements io.Writer
type KlogWriter struct {
}

func (*KlogWriter) Write(p []byte) (int, error) {
	klog.V(5).Infof(string(p))
	return len(p), nil
}

const databaseAlias = "default"

func NewDatabaseConnection(e bool, t, h string, p int, u, s, d, c string) (*DatabaseConnection, error) {
	cfg := DatabaseConfig{e, t, h, p, u, s, d, c}
	return newDatabaseConnection(cfg)
}

type DatabaseConfig struct {
	Enabled  bool
	Type     string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Schema   string
}

var c = make(chan struct{})

func newDatabaseConnection(cfg DatabaseConfig) (*DatabaseConnection, error) {
	//panic when called twice
	close(c)
	if !cfg.Enabled {
		return &DatabaseConnection{enabled: false}, nil
	}
	orm.DebugLog = orm.NewLog(new(KlogWriter))
	orm.Debug = true
	var err error

	switch cfg.Type {
	case "mysql":
		orm.RegisterDriver("mysql", orm.DRMySQL)
		err = orm.RegisterDataBase(databaseAlias, "mysql", cfg.Username+":"+cfg.Password+"@tcp("+cfg.Host+":"+strconv.Itoa(cfg.Port)+")/"+cfg.Database)
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
	orm.RegisterModel(new(model.Application), new(model.UserRelation), new(model.Task), new(model.TbDagRun), new(model.TbMetadata),
		new(model.Relation), new(model.Api), new(model.Serviceunit))
	if err = orm.RunSyncdb("default", false, true); err != nil {
		return nil, fmt.Errorf("cannot sync database: %+v", err)
	}
	return &DatabaseConnection{orm.NewOrm(), &sync.Mutex{}, cfg.Enabled}, nil
}

func (d *DatabaseConnection) Enabled() bool {
	return d.enabled
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

func (d *DatabaseConnection) Rollback() error {
	defer d.mtx.Unlock()
	return d.Ormer.Rollback()
}

func (d *DatabaseConnection) query(uid string, md model.Table, conditions []model.Condition, result interface{}) (err error) {
	var sql string
	var values []interface{}
	and := false
	if len(uid) > 0 {
		sqlTpl := `SELECT * FROM %s WHERE id IN (SELECT resource_id FROM user_relation WHERE resource_type = "%s" AND user_id = "%s")`
		sql = fmt.Sprintf(sqlTpl, md.TableName(), md.ResourceType(), uid)
		//sql = fmt.Sprintf(sqlTpl, md.TableName())
		and = true
	} else {
		sql = `SELECT * FROM ` + md.TableName()
		values = make([]interface{}, len(conditions))
	}
	for i, c := range conditions {
		var k, o, v string
		k = c.Field
		v = c.Value
		switch c.Operator {
		case model.Equals:
			o = "="
		case model.LessThan:
			o = "<"
		case model.NotLessThan:
			o = ">="
		case model.MoreThan:
			o = ">"
		case model.NotMoreThan:
			o = "<="
		case model.Like:
			o = "LIKE"
			v = `%%` + v + `%%`
		case model.In:
			o = "IN"
			v = `(` + v + `)`
		}
		if and {
			sql = sql + " AND "
		} else {
			sql = sql + " WHERE "
		}
		if len(uid) > 0 {
			sql = sql + k + " " + o + " \"" + v + "\""
		} else {
			sql = sql + k + " " + o + " ?"
			values[i] = v
		}
		and = true
	}
	_, err = d.Raw(sql, values...).QueryRows(result)
	return
}

func (d *DatabaseConnection) AddObject(o model.Table, us []model.UserRelation, count int, rl []model.Relation, rcount int) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Insert(o)
	if err != nil {
		d.Rollback()
		return err
	}
	if us != nil && len(us) > 0 {
		_, err = d.InsertMulti(count, us)
		if err != nil {
			d.Rollback()
			return err
		}
	}
	if rl != nil && len(rl) > 0 {
		_, err = d.InsertMulti(rcount, rl)
		if err != nil {
			d.Rollback()
			return err
		}
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) UpdateObject(o model.Table, us []model.UserRelation, count int, rl []model.Relation, rcount int) (err error) {
	if err := d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	if _, err := d.Update(o); err != nil {
		d.Rollback()
		return err
	}
	if us != nil && len(us) > 0 {
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
	if rl != nil && len(rl) > 0 {
		rl := []model.Relation{}
		if _, err := d.QueryTable("Relation").Filter("SourceId", o.ResourceType()).Filter("SourceType", o.ResourceType()).All(&rl); err != nil {
			d.Rollback()
			return err
		}
		for _, u := range rl {
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
	if err != nil {
		d.Rollback()
		return err
	}

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

	rl := []model.Relation{}
	if _, err := d.QueryTable("Relation").Filter("SourceId", o.ResourceId()).Filter("SourceType", o.ResourceType()).All(&rl); err != nil {
		d.Rollback()
		return err
	}
	for _, r := range rl {
		if _, err := d.Delete(&r); err != nil {
			d.Rollback()
			return err
		}
	}

	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) AddApplication(obj interface{}) error {
	o, us, rl, err := model.ApplicationGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get application from obj error: %+v", err)
	}
	return d.AddObject(&o, us, len(us), rl, len(rl))
}

func (d *DatabaseConnection) UpdateApplication(old, obj interface{}) error {
	_, ous, _, err := model.ApplicationGetFromObject(old)
	if err != nil {
		return fmt.Errorf("get users from obj error: %+v", err)
	}
	o, us, rl, err := model.ApplicationGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get application from obj error: %+v", err)
	}
	if model.Equal(ous, us) {
		return d.UpdateObject(&o, nil, 0, rl, len(rl))
	}

	return d.UpdateObject(&o, us, len(us), rl, len(rl))
}

func (d *DatabaseConnection) DeleteApplication(obj interface{}) error {
	o, _, _, err := model.ApplicationGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get application from obj error: %+v", err)
	}
	return d.DeleteObject(&o)
}

func (d *DatabaseConnection) QueryApplication(uid string, md *model.Application) (result []model.Application, err error) {
	conditions := make([]model.Condition, 0)
	if md == nil {
		return nil, fmt.Errorf("model is null")
	}
	if len(md.Namespace) == 0 {
		return nil, fmt.Errorf("namespace not set in model")
	}
	conditions = append(conditions, model.Condition{"namespace", model.Equals, md.Namespace})
	if len(md.Group) > 0 {
		conditions = append(conditions, model.Condition{"group", model.Equals, md.Group})
	}
	if len(md.Name) > 0 {
		conditions = append(conditions, model.Condition{"name", model.Like, md.Name})
	}
	if len(md.Status) > 0 {
		conditions = append(conditions, model.Condition{"status", model.Equals, md.Status})
	}

	err = d.query(uid, md, conditions, &result)
	return
}

func (d *DatabaseConnection) AddApi(obj interface{}) error {
	o, us, err := model.ApiGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get api from obj error: %+v", err)
	}
	return d.AddObject(&o, us, len(us), nil, 0)
}

func (d *DatabaseConnection) UpdateApi(old, obj interface{}) error {
	_, ous, err := model.ApiGetFromObject(old)
	if err != nil {
		return fmt.Errorf("get users from obj error: %+v", err)
	}
	o, us, err := model.ApiGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get api from obj error: %+v", err)
	}
	if model.Equal(ous, us) {
		return d.UpdateObject(&o, nil, 0, nil, 0)
	}

	return d.UpdateObject(&o, us, len(us), nil, 0)
}

func (d *DatabaseConnection) DeleteApi(obj interface{}) error {
	o, _, err := model.ApiGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get api from obj error: %+v", err)
	}
	return d.DeleteObject(&o)
}

func (d *DatabaseConnection) AddServiceunit(obj interface{}) error {
	o, us, err := model.ServiceunitGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get serviceunit from obj error: %+v", err)
	}
	return d.AddObject(&o, us, len(us), nil, 0)
}

func (d *DatabaseConnection) UpdateServiceunit(old, obj interface{}) error {
	_, ous, err := model.ServiceunitGetFromObject(old)
	if err != nil {
		return fmt.Errorf("get users from obj error: %+v", err)
	}
	o, us, err := model.ServiceunitGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get serviceunit from obj error: %+v", err)
	}
	if model.Equal(ous, us) {
		return d.UpdateObject(&o, nil, 0, nil, 0)
	}

	return d.UpdateObject(&o, us, len(us), nil, 0)
}

func (d *DatabaseConnection) DeleteServiceunit(obj interface{}) error {
	o, _, err := model.ServiceunitGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get serviceunit from obj error: %+v", err)
	}
	return d.DeleteObject(&o)
}
