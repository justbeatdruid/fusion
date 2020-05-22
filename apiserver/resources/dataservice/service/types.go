package service

import (
	"encoding/json"
	"time"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/names"
)

//SchedualPlan ...
type SchedualPlan struct {
	QuartzCron           bool
	QuartzCronExpression string
	TimeUnit             string
	SchedualPeriod       int
	StartTime            time.Time
	Description          string
}

//DataSource ...
type DataSource struct {
	Type         string
	RelationalDb RelationalDbConfig
}

//RelationalDbConfig ...
type RelationalDbConfig struct {
	Name                  string
	ExecSql               string
	SourceTable           string
	SortField             string
	SortMode              string
	IncrementalMigration  bool
	TimeZone              Zone
	Timestamp             string
	TimestampInitialValue time.Time
	TimeCompensation      int
	Condition             []ConditionConfig
}

//Zone ...
type Zone struct {
	Name  string
	Offet int
}

//ConditionConfig ...
type ConditionConfig struct {
	Relation  string
	LeftField string
	Operation string
	Right     interface{}
}

//DataTarget ...
type DataTarget struct {
	Type               string
	RelationalDbTarget RelationalDbTarget
}

//RelationalDbTarget ...
type RelationalDbTarget struct {
	Name            string
	TargetTable     string
	MappingRelation []MappingRelationConfig
}

//MappingRelationConfig ...
type MappingRelationConfig struct {
	SourceField  string
	SourceType   string
	SourceLength int
	TargetField  string
	TargetType   string
	TargetLength int
}

// Dataservice ...
type Dataservice struct {
	ID                 int    `json:"Id"`
	DagID              string `json:"DagId"`
	Name               string
	Description        string
	Namespace          string
	Type               string //realtime or periodic
	SchedualPlanConfig SchedualPlan
	DataSourceConfig   DataSource
	DataTargetConfig   DataTarget
	CreatedAt          time.Time
	Status             string
}

// ToAPI  only used in creation options
func ToAPI(ds *model.Task) *Dataservice {
	crd := &Dataservice{
		ID:          ds.Id,
		DagID:       ds.DagId,
		Name:        ds.Name,
		Description: ds.Description,
		Namespace:   "default",
		Type:        ds.Type,
		CreatedAt:   ds.CreatedTime,
	}
	json.Unmarshal([]byte(ds.SchedualPlan), &crd.SchedualPlanConfig)
	json.Unmarshal([]byte(ds.DataSourceConfig), &crd.DataSourceConfig)
	json.Unmarshal([]byte(ds.DataTargetConfig), &crd.DataTargetConfig)

	return crd
}

//ToModel ...
func ToModel(obj *Dataservice) *model.Task {
	ds := &model.Task{
		DagId: names.NewID(),
		// Namespace: obj.ObjectMeta.Namespace,
		Name:        obj.Name,
		Description: obj.Description,
		Type:        obj.Type,
		CreatedTime: time.Now(),
	}
	plan, _ := json.Marshal(obj.SchedualPlanConfig)
	ds.SchedualPlan = string(plan)
	dataSourceConfig, _ := json.Marshal(obj.DataSourceConfig)
	ds.DataSourceConfig = string(dataSourceConfig)
	dataTargetConfig, _ := json.Marshal(obj.DataTargetConfig)
	ds.DataTargetConfig = string(dataTargetConfig)

	return ds
}

//ToListAPI ...
func ToListAPI(items []model.Task) []*Dataservice {
	ds := make([]*Dataservice, len(items))
	for i := range items {
		ds[i] = ToAPI(&items[i])
	}
	return ds
}

//FlinkxReq ...
type FlinkxReq struct {
	Job Job `json:"job"`
}

//NewFlinkxReq ...
func NewFlinkxReq() *FlinkxReq {
	return &FlinkxReq{
		Job: Job{
			Setting: Setting{
				Speed: Speed{
					Channel: 3,
					Bytes:   0,
				},
				ErrorLimit: ErrorLimit{
					Record:     10000,
					Percentage: 100,
				},
				Dirty: map[string]interface{}{"path": "/tmp"},
			},
			Content: []Content{
				Content{
					Reader: Reader{
						Name: "mysqlreader",
						Parameter: Parameter{
							UserName: "root",
							Password: "12345",
							Column:   []string{"id", "name"},
							Where:    "id>1",
							Connection: []Connection{Connection{
								Table:   []string{"testreader"},
								JdbcURL: []string{"jdbc:mysql://nlstore-mysql:3306/testreader"},
							}},
							SplitPk: "id",
						},
					},
					Writer: Writer{
						Name: "mysqlwriter",
						Parameter: WriteParameter{
							WriteMode: "insert",
							UserName:  "root",
							Password:  "123456",
							Column:    []string{"id", "name"},
							BatchSize: 1,
							Session:   []string{"set session sql_mode='ANSI'"},
							Connection: []Connection{Connection{
								Table:   []string{"testwriter"},
								JdbcURL: []string{"jdbc:mysql://nlstore-mysql:3306/testwriter"},
							}},
						},
					},
				},
			},
		},
	}
}

//Job ...
type Job struct {
	Setting Setting   `json:"setting"`
	Content []Content `json:"content"`
}

//Setting ...
type Setting struct {
	Speed      Speed       `json:"speed"`
	ErrorLimit ErrorLimit  `json:"errorLimit"`
	Dirty      interface{} `json:"dirty"`
}

//Speed ...
type Speed struct {
	Channel int `json:"channel"`
	Bytes   int `json:"bytes"`
}

//ErrorLimit ...
type ErrorLimit struct {
	Record     int `json:"record"`
	Percentage int `json:"percentage"`
}

//Content ...
type Content struct {
	Reader Reader `json:"reader"`
	Writer Writer `json:"writer"`
}

//Reader ...
type Reader struct {
	Name      string    `json:"Id"`
	Parameter Parameter `json:"parameter"`
}

//Parameter ...
type Parameter struct {
	UserName   string       `json:"username"`
	Password   string       `json:"password"`
	Column     []string     `json:"column"`
	Where      string       `json:"where"`
	Connection []Connection `json:"connection"`
	SplitPk    string       `json:"splitPk"`
}

//Connection ...
type Connection struct {
	Table   []string `json:"table"`
	JdbcURL []string `json:"jdbcUrl"`
}

//Writer ...
type Writer struct {
	Name      string         `json:"Id"`
	Parameter WriteParameter `json:"parameter"`
}

//WriteParameter ...
type WriteParameter struct {
	WriteMode  string       `json:"writeMode"`
	UserName   string       `json:"username"`
	Password   string       `json:"password"`
	Column     []string     `json:"column"`
	BatchSize  int          `json:"batchSize"`
	Session    []string     `json:"session"`
	Connection []Connection `json:"connection"`
}

// func (ds *Dataservice) Validate() error {
// 	for k, v := ranAPIge map[string]string{
// 		"type": string(ds.Type),
// 	} {
// 		if len(v) == 0 {
// 			return fmt.Errorf("%s is null", k)
// 		}
// 	}
// 	switch ds.Type {
// 	case v1.Realtime:
// 	case v1.Periodic:
// 		if err := v1.ValidateCronConfig(ds.CronConfig); err != nil {
// 			return fmt.Errorf("cron config error: %+v", err)
// 		}
// 	default:
// 		return fmt.Errorf("wrong task type: %s", ds.Type)
// 	}
// 	ds.ID = names.NewID()
// 	return nil
// }
