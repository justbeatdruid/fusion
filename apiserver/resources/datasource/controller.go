package datasource

import (
	"database/sql"
	"fmt"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
	"strings"

	"github.com/chinamobile/nlpt/apiserver/resources/datasource/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/emicklei/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.DatasourceConfig.Supported),
	}
}

type Wrapped struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *service.Datasource `json:"data,omitempty"`
}
type QueryDataResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"queryData"`
}
type QueryMysqlDataResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    []map[string]string `json:"queryData"`
}
type CreateResponse = Wrapped
type UpdateResponse = Wrapped
type CreateRequest = Wrapped
type UpdateRequest = Wrapped
type DeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type PingResponse = DeleteResponse

func (c *controller) CreateDatasource(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if db, err := c.service.CreateDatasource(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
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
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if db, err := c.service.UpdateDatasource(body.Data); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    2,
			Message: fmt.Errorf("update database error: %+v", err).Error(),
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
	if db, err := c.service.GetDatasource(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
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
	if err := c.service.DeleteDatasource(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:    0,
			Message: "",
		}
	}
}

type DataSourceList []*service.Datasource

func (dsl DataSourceList) Length() int {
	return len(dsl)
}

func (dsl DataSourceList) GetItem(i int) (interface{}, error) {
	if i >= len(dsl) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return dsl[i], nil
}
func (c *controller) ListDatasource(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	if db, err := c.service.ListDatasource(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var dbs DataSourceList = db
		pageStruct, err := util.PageWrap(dbs, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    1,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:    0,
			Message: "success",
			Data:    pageStruct,
		}
	}
}
func (c *controller) getDataByApi(req *restful.Request) (int, *QueryDataResponse) {
	apiId := req.PathParameter("apiId")
	//todo Acquisition parameters in the request body（Provisional use this method）
	parameters, err := req.BodyParameter("params")
	if err != nil {
		return http.StatusInternalServerError, &QueryDataResponse{
			Code:    1,
			Message: fmt.Errorf("get parameters error: %+v", err).Error(),
		}
	}
	result, err := c.service.GetDataSourceByApiId(apiId, parameters)
	if err != nil {
		return http.StatusInternalServerError, &QueryDataResponse{
			Code:    1,
			Message: fmt.Errorf("query data error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &QueryDataResponse{
		Code: 0,
		Data: result,
	}
}
func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}

/**链接MySQL
 */
func (c *controller) ConnectMysql(req *restful.Request) (int, interface{}) {
	connect := &service.Connect{}
	req.ReadEntity(connect)
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
		fmt.Println("open DB err")
		return http.StatusInternalServerError, &QueryMysqlDataResponse{
			Code:    1,
			Message: "open DB err",
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
			Code:    1,
			Message: "open database fail",
		}
	}
	var querySql string
	if len(connect.TableName) != 0 {
		//查询表字段名称，字段描述，字段类型
		querySql = "select COLUMN_NAME '字段名称',COLUMN_TYPE '字段类型长度',IF(EXTRA='auto_increment',CONCAT(COLUMN_KEY," +
			"'(', IF(EXTRA='auto_increment','自增长',EXTRA),')'),COLUMN_KEY) '主外键',IS_NULLABLE '空标识',COLUMN_COMMENT " +
			"'字段说明' from information_schema.columns where table_name='" + connect.TableName + "'" +
			" and table_schema='" + connect.DbName + "'"
	} else {
		//查询数据库所有表名
		querySql = "SELECT distinct TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '" + connect.DbName + "'"
	}
	fmt.Println("connnect success")
	data, err := service.GetMySQLDbData(db, querySql)
	if err != nil {
		return http.StatusInternalServerError, &QueryMysqlDataResponse{
			Code:    1,
			Message: "deal data fail",
		}
	}
	return http.StatusOK, &QueryMysqlDataResponse{
		Code: 0,
		Data: data,
	}
}
