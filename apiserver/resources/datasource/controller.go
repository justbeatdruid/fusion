package datasource

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chinamobile/nlpt/apiserver/resources/datasource/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errCode map[string]string
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.DatasourceConfig.Supported, cfg.DataserviceConnector, cfg.TenantEnabled),
		cfg.LocalConfig.DataSource,
	}
}

type Wrapped struct {
	Code      int                 `json:"code"`
	ErrorCode string              `json:"errorCode"`
	Detail    string              `json:"detail"`
	Message   string              `json:"message"`
	Data      *service.Datasource `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.Datasource `json:"data,omitempty"`
}

type CreateRequest = RequestWrapped
type UpdateRequest = RequestWrapped
type PingRequest = RequestWrapped
type QueryRequest struct {
	SQL string `json:"sql"`
}
type QueryResponse struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type UpdateResponse = Wrapped
type DeleteResponse Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type Unstructured struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

type StatisticsResponse = struct {
	Code      int                `json:"code"`
	ErrorCode string             `json:"errorCode"`
	Message   string             `json:"message"`
	Data      service.Statistics `json:"data"`
	Detail    string             `json:"detail"`
}

func (c *controller) CreateDatasource(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if db, err := c.service.CreateDatasource(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   2,
			Detail: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) UpdateDatasource(req *restful.Request) (int, *UpdateResponse) {
	body := &UpdateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:   1,
			Detail: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	if len(body.Data.ID) == 0 {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:   1,
			Detail: "read entity error: id in body is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	if db, err := c.service.UpdateDatasource(body.Data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:   2,
			Detail: fmt.Errorf("update database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) GetDatasource(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	if db, err := c.service.GetDatasource(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:   1,
			Detail: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) DeleteDatasource(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	if err := c.service.DeleteDatasource(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:   1,
			Detail: fmt.Errorf("delete database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:   0,
			Detail: "",
		}
	}
}

type DataSourceList []*service.Datasource

func (dsl DataSourceList) Len() int {
	return len(dsl)
}

func (dsl DataSourceList) GetItem(i int) (interface{}, error) {
	if i >= len(dsl) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return dsl[i], nil
}

func (dsl DataSourceList) Less(i, j int) bool {
	return dsl[i].CreatedAt.Time.After(dsl[j].CreatedAt.Time)
}

func (dsl DataSourceList) Swap(i, j int) {
	dsl[i], dsl[j] = dsl[j], dsl[i]
}

func (c *controller) ListDatasource(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	typpe := req.QueryParameter("type")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	if db, err := c.service.ListDatasource(util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace),
		util.WithNameLike(name), util.WithType(typpe)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:   1,
			Detail: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var dbs DataSourceList = db
		pageStruct, err := util.PageWrap(dbs, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:   1,
				Detail: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:   0,
			Detail: "success",
			Data:   pageStruct,
		}
	}
}

func (c *controller) Ping(req *restful.Request) (int, *PingResponse) {
	body := &PingRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &PingResponse{
			Code:   1,
			Detail: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &PingResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &PingResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if err = c.service.Ping(body.Data); err != nil {
		return http.StatusInternalServerError, &PingResponse{
			Code:   1,
			Detail: fmt.Errorf("ping database error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &PingResponse{
		Code: 0,
	}
}

func (c *controller) Query(req *restful.Request) (int, *QueryResponse) {
	id := req.PathParameter("id")
	body := &QueryRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &QueryResponse{
			Code:   1,
			Detail: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.SQL == "" {
		return http.StatusInternalServerError, &QueryResponse{
			Code:   1,
			Detail: "read entity error: sql is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &QueryResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	res, err := c.service.Query(id, body.SQL, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &QueryResponse{
			Code:   1,
			Detail: fmt.Errorf("exec sql error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &QueryResponse{
		Code: 0,
		Data: res,
	}
}

func (c *controller) GetTables(req *restful.Request) (int, *Unstructured) {
	id := req.PathParameter("id")
	associationID := req.QueryParameter("association")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &Unstructured{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	result, err := c.service.GetTables(id, associationID, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &Unstructured{
			Code:   1,
			Detail: fmt.Errorf("get tables error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &Unstructured{
		Code: 0,
		Data: result,
	}
}

func (c *controller) GetTable(req *restful.Request) (int, *Unstructured) {
	id := req.PathParameter("id")
	table := req.PathParameter("table")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &Unstructured{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	result, err := c.service.GetTable(id, table, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &Unstructured{
			Code:   1,
			Detail: fmt.Errorf("get tables error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &Unstructured{
		Code: 0,
		Data: result,
	}
}

func (c *controller) GetFields(req *restful.Request) (int, *Unstructured) {
	id := req.PathParameter("id")
	// for quering RDB fields, pass table name in query parameter "table"
	// for quering datawarehouse fields, pass table ID in parameter
	table := req.PathParameter("table")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &Unstructured{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	result, err := c.service.GetFields(id, table, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &Unstructured{
			Code:   1,
			Detail: fmt.Errorf("get fields: %+v", err).Error(),
		}
	}
	return http.StatusOK, &Unstructured{
		Code: 0,
		Data: result,
	}
}

func (c *controller) GetField(req *restful.Request) (int, *Unstructured) {
	id := req.PathParameter("id")
	// for quering RDB fields, pass table name in query parameter "table"
	// for quering datawarehouse fields, pass table ID in parameter
	table := req.PathParameter("table")
	field := req.PathParameter("field")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "006000005"
		return http.StatusInternalServerError, &Unstructured{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	result, err := c.service.GetField(id, table, field, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &Unstructured{
			Code:   1,
			Detail: fmt.Errorf("get fields: %+v", err).Error(),
		}
	}
	return http.StatusOK, &Unstructured{
		Code: 0,
		Data: result,
	}
}

/*
func (c *controller) getDataByApi(req *restful.Request) (int, *QueryDataResponse) {
	apiId := req.PathParameter("apiId")
	//todo Acquisition parameters in the request body（Provisional use this method）
	parameters, err := req.BodyParameter("params")
	if err != nil {
		return http.StatusInternalServerError, &QueryDataResponse{
			Code:   1,
			Detail: fmt.Errorf("get parameters error: %+v", err).Error(),
		}
	}
	result, err := c.service.GetDataSourceByApiId(apiId, parameters)
	if err != nil {
		return http.StatusInternalServerError, &QueryDataResponse{
			Code:   1,
			Detail: fmt.Errorf("query data error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &QueryDataResponse{
		Code: 0,
		Data: result,
	}
}
*/

func (c *controller) DoStatistics(req *restful.Request) (int, *StatisticsResponse) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: "002000003",
			Message:   c.errCode["002000003"],
			Detail:    "auth model error",
		}
	}
	appList, err := c.service.List(util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: "002000008",
			Message:   c.errCode["002000008"],
			Detail:    fmt.Sprintf("do statistics on apps error, %+v", err),
		}
	}

	data := service.Statistics{}
	data.Total = len(appList.Items)
	data.Increment, data.Percentage = c.CountAppsIncrement(appList.Items)
	return http.StatusOK, &StatisticsResponse{
		Code:      0,
		ErrorCode: "",
		Message:   "",
		Data:      data,
		Detail:    "do statistics successfully",
	}
}

func (c *controller) CountAppsIncrement(dss []v1.Datasource) (int, string) {
	var increment int
	var percentage string
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	for _, t := range dss {
		createTime := util.NewTime(t.ObjectMeta.CreationTimestamp.Time)
		if createTime.Unix() < end.Unix() && createTime.Unix() >= start.Unix() {
			increment++
		}
	}
	total := len(dss)
	var pre float64
	if total > 0 {
		pre = float64(increment) / float64(total) * 100
	}
	percentage = fmt.Sprintf("%.0f%s", pre, "%")

	return increment, percentage
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}

/**链接MySQL
func getParams(req *restful.Request) *service.Connect {
	connect := &service.Connect{}
	connect.UserName = req.QueryParameter("UserName")
	connect.Password = req.QueryParameter("Password")
	connect.Ip = req.QueryParameter("Ip")
	connect.Port = req.QueryParameter("Port")
	connect.DBName = req.QueryParameter("DBName")
	connect.TableName = req.QueryParameter("TableName")
	queryCondition := make(map[string]string)
	err := json.Unmarshal([]byte(req.QueryParameter("QueryCondition")), &queryCondition)
	if err != nil {
		connect.QueryCondition = nil
	} else {
		connect.QueryCondition = queryCondition
	}
	connect.QType = req.QueryParameter("QType")
	return connect
}

func (c *controller) ConnectMysql(req *restful.Request) (int, interface{}) {
	connect := getParams(req)
	if len(connect.QType) == 0 {
		if connect.QType == "3" && connect.QueryCondition == nil {
			fmt.Println("parameter error,条件为空")
			return http.StatusInternalServerError, &QueryMysqlDataResponse{
				Code:   1,
				Detail: "parameter error,条件不能为空",
			}
		}
		fmt.Println("parameter error,qType 为空")
		return http.StatusInternalServerError, &QueryMysqlDataResponse{
			Code:   1,
			Detail: "parameter error,qType 不能为空",
		}
	}
	buildPath := strings.Builder{}
	buildPath.WriteString(connect.UserName)
	buildPath.WriteString(":")
	buildPath.WriteString(connect.Password)
	buildPath.WriteString("@tcp(")
	buildPath.WriteString(connect.Ip)
	buildPath.WriteString(":")
	buildPath.WriteString(connect.Port)
	buildPath.WriteString(")/")
	buildPath.WriteString(connect.DBName)
	path := buildPath.String()
	db, err := sql.Open("mysql", path)
	if err != nil {
		fmt.Println("open DB err", err)
		fmt.Println(err)
		return http.StatusInternalServerError, &QueryMysqlDataResponse{
			Code:   1,
			Detail: "open DB err，path :" + path,
		}
	}
	//设置数据库最大连接数
	db.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	db.SetMaxIdleConns(10)
	//验证连接
	if err := db.Ping(); err != nil {
		fmt.Println("open database fail")
		return http.StatusInternalServerError, &QueryMysqlDataResponse{
			Code:   1,
			Detail: "open database fail",
		}
	}
	var querySql string

	if connect.QType == "1" {
		//查询表字段名称，字段描述，字段类型
		querySql = "select COLUMN_NAME '字段名称',COLUMN_TYPE '字段类型长度',IF(EXTRA='auto_increment',CONCAT(COLUMN_KEY," +
			"'(', IF(EXTRA='auto_increment','自增长',EXTRA),')'),COLUMN_KEY) '主外键',IS_NULLABLE '空标识',COLUMN_COMMENT " +
			"'字段说明' from information_schema.columns where table_name='" + connect.TableName + "'" +
			" and table_schema='" + connect.DBName + "'"
	} else if connect.QType == "2" {
		//查询数据库所有表名
		querySql = "SELECT distinct table_name FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '" + connect.DBName + "'"
	} else {
		//查询数据表数据
		querySqls := strings.Builder{}
		querySqls.WriteString("SELECT * FROM " + "`" + connect.DBName + "`." + "`" + connect.TableName + "`")
		if connect.QueryCondition != nil && len(connect.QueryCondition) > 0 {
			querySqls.WriteString("where ")
			for k, v := range connect.QueryCondition {
				querySqls.WriteString(k + "=" + "'" + v + "'" + " and ")
			}
			querySql = querySqls.String()            //拼接的sql语句转成字符串
			querySql = querySql[0 : len(querySql)-4] //截取最后三个字符“and”
		} else {
			querySql = querySqls.String() //拼接的sql语句转成字符串
		}
		fmt.Println("querySql: " + querySql)

	}
	data, err := service.GetMySQLDbData(db, querySql)
	if err != nil {
		return http.StatusInternalServerError, &QueryMysqlDataResponse{
			Code:   1,
			Detail: "deal data fail",
		}
	}
	return http.StatusOK, &QueryMysqlDataResponse{
		Code:   0,
		Detail: "查询成功",
		Data:   data,
	}
}
*/
