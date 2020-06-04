package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/util"
	"k8s.io/klog"
)

//SchedualPlan ...
type SchedualPlan struct {
	QuartzCron           bool
	QuartzCronExpression string
	TimeUnit             string
	SchedualPeriod       int
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
	SourceID              string
	ExecSql               ExecSql
	SourceTable           string
	SortField             string
	SortMode              string
	IncrementalMigration  bool
	TimeZone              Zone
	Timestamp             string
	TimestampInitialValue string
	TimeCompensation      int
	Condition             []ConditionConfig
}

type ExecSql struct {
	ExecSqlFlag  bool
	ExecSqlWords string
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
	TargetID        string
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
	StartTime          string
	SchedualPlanConfig SchedualPlan
	DataSourceConfig   DataSource
	DataTargetConfig   DataTarget
	CreatedAt          string
	ExecDate           string
	Status             bool
	DagInfo            DagInfo
	DagStatus          int
}

// OperationReq ...
type OperationReq struct {
	Operation string
	DagID     []string
}

//Validate ...
func (operation *OperationReq) Validate() error {

	if operation.Operation != "stop" && operation.Operation != "open" && operation.Operation != "delete" {
		return fmt.Errorf("Operation is error value,The value range of type is stop , open or delete,Operation:%v", operation.Operation)
	}

	return nil
}

func (ds *Dataservice) Validate(service *Service, opts ...util.OpOption) error {
	if len(ds.Name) == 0 {
		return fmt.Errorf("name is null,name %v", ds.Name)
	}
	if ds.Type != "realtime" && ds.Type != "periodic" {
		return fmt.Errorf("type is error value,The value range of type is realtime or periodic,type:%v", ds.Type)
	}
	if ds.DataSourceConfig.RelationalDb.SourceID == "" || ds.DataTargetConfig.RelationalDbTarget.TargetID == "" {
		return fmt.Errorf("SourceID or TargetID is error value,SourceID:%v,TargetID:%v", ds.DataSourceConfig.RelationalDb.SourceID, ds.DataTargetConfig.RelationalDbTarget.TargetID)
	}
	dataSource, _, err := service.GetDataSource(ds.DataSourceConfig.RelationalDb.SourceID, opts...)
	klog.Errorf("dataSource:%v,err:%v", dataSource, err)
	if err != nil {
		return fmt.Errorf("find dataSource failed by SourceID ,sourceID:%v.", ds.DataSourceConfig.RelationalDb.SourceID)
	}
	dataSource, _, err = service.GetDataSource(ds.DataTargetConfig.RelationalDbTarget.TargetID, opts...)
	klog.Errorf("dataSource:%v,err:%v", dataSource, err)
	if err != nil {
		return fmt.Errorf("find dataSource failed by TargetID ,sourceID:%v.", ds.DataTargetConfig.RelationalDbTarget.TargetID)
	}
	if _, err = time.Parse(TimeStr, ds.StartTime); err != nil {
		return fmt.Errorf("StartTime error,:err:%v,startTime:%v", err, ds.StartTime)
	}
	if _, err = time.Parse(TimeStr, ds.DataSourceConfig.RelationalDb.TimestampInitialValue); err != nil {
		return fmt.Errorf("StartTime error,:err:%v,TimestampInitialValue:%v", err, ds.DataSourceConfig.RelationalDb.TimestampInitialValue)
	}
	return nil
}

// DagStatus ...
type DagInfo struct {
	ErrorMessage    string `json:"errorMessage"`
	NumWrite        int64  `json:"numWrite"`
	SourceTableName string `json:"sourceTableName"`
	ByteRead        int64  `json:"byteRead"`
	CostTime        string `json:"costTime"`
	NumRead         int64  `json:"NumRead"` //realtime or periodic
	NumFilter       int64  `json:"numFilter"`
	TargetTableName string `json:"targetTableName"`
	ByteWrite       int64  `json:"byteWrite"`
	NErrors         int64  `json:"nErrors"`
}

const (
	Maxlimit      = 100
	DefaultOffset = 0
	DefaultPage   = 1
	TimeStr       = "2006-01-02 15:04:05"
)

// ToAPI  only used in creation options
func ToAPI(ds *model.Task, service *Service, opts ...util.OpOption) *Dataservice {
	apiTask := &Dataservice{
		ID:          ds.Id,
		DagID:       ds.DagId,
		Name:        ds.Name,
		Description: ds.Description,
		Namespace:   "default",
		Type:        ds.Type,
		Status:      ds.Status,
		StartTime:   ds.StartTime.Format(TimeStr),
		CreatedAt:   ds.CreatedTime.Format(TimeStr),
	}
	json.Unmarshal([]byte(ds.SchedualPlan), &apiTask.SchedualPlanConfig)
	json.Unmarshal([]byte(ds.DataSourceConfig), &apiTask.DataSourceConfig)
	json.Unmarshal([]byte(ds.DataTargetConfig), &apiTask.DataTargetConfig)
	_, apiTask.DataTargetConfig.RelationalDbTarget.Name, _ = service.GetDataSource(apiTask.DataTargetConfig.RelationalDbTarget.TargetID, opts...)
	_, apiTask.DataSourceConfig.RelationalDb.Name, _ = service.GetDataSource(apiTask.DataSourceConfig.RelationalDb.SourceID, opts...)
	dagRun, num, err := model.GetTbDagRun(apiTask.DagID)
	if num > 0 && err == nil {
		apiTask.DagStatus = dagRun[0].DagStatus
		apiTask.ExecDate = dagRun[0].ExecDate.Format(TimeStr)
		json.Unmarshal([]byte(dagRun[0].Remark), &apiTask.DagInfo)

	}
	return apiTask
}

//ToModel ...
func ToModel(obj *Dataservice) *model.Task {
	ds := &model.Task{
		//DagId: names.NewID(),
		// Namespace: obj.ObjectMeta.Namespace,
		Name:        obj.Name,
		Description: obj.Description,
		Type:        obj.Type,
		//StartTime:   obj.StartTime,
		CreatedTime: time.Now(),
	}
	ds.StartTime, _ = time.Parse(TimeStr, obj.StartTime)
	plan, _ := json.Marshal(obj.SchedualPlanConfig)
	ds.SchedualPlan = string(plan)
	dataSourceConfig, _ := json.Marshal(obj.DataSourceConfig)
	ds.DataSourceConfig = string(dataSourceConfig)
	dataTargetConfig, _ := json.Marshal(obj.DataTargetConfig)
	ds.DataTargetConfig = string(dataTargetConfig)

	return ds
}

//ToListAPI ...
func ToListAPI(items []model.Task, service *Service, opts ...util.OpOption) []*Dataservice {
	ds := make([]*Dataservice, len(items))
	for i := range items {
		ds[i] = ToAPI(&items[i], service, opts...)
	}
	return ds
}

func (s *Service) assignment(taskDb *model.Task, data map[string]interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var source Dataservice
	if err = json.Unmarshal(b, &source); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	if _, ok := data["Name"]; ok {

		taskDb.Name = source.Name
	}
	if _, ok := data["Description"]; ok {
		taskDb.Description = source.Description
	}

	if _, ok := data["Type"]; ok {
		taskDb.Type = source.Type
	}

	if _, ok := data["SchedualPlanConfig"]; ok {
		plan, _ := json.Marshal(source.SchedualPlanConfig)
		taskDb.SchedualPlan = string(plan)
	}

	if _, ok := data["DataSourceConfig"]; ok {
		dataSourceConfig, _ := json.Marshal(source.DataSourceConfig)
		taskDb.DataSourceConfig = string(dataSourceConfig)
	}

	if _, ok := data["DataTargetConfig"]; ok {
		dataTargetConfig, _ := json.Marshal(source.DataTargetConfig)
		taskDb.DataTargetConfig = string(dataTargetConfig)
	}

	return nil
}

//FlinkxReq ...
type FlinkxReq struct {
	Job Job `json:"job"`
}

//NewFlinkxReq ...
func NewFlinkxReq() FlinkxReq {
	return FlinkxReq{
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
							//Column:   []string{"id", "name"},
							Column: []string{},
							Where:  "id>1",
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
							//Column:    []string{"id", "name"},
							Column:    []string{},
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
	Name      string    `json:"name"`
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
	Name      string         `json:"name"`
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
