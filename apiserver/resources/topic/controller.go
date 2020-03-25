package topic

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	tperror "github.com/chinamobile/nlpt/apiserver/resources/topic/error"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/parser"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/chinamobile/nlpt/pkg/go-restful/log"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
	"strconv"
	"strings"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.TopicConfig),
		cfg.LocalConfig,
	}
}

type Wrapped struct {
	Code      int            `json:"code"`
	ErrorCode string         `json:"errorCode"`
	Message   string         `json:"message"`
	Data      *service.Topic `json:"data,omitempty"`
	Detail    string         `json:"detail"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped
type GrantResponse = Wrapped

/*type DeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}*/
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Detail    string      `json:"detail"`
}
type MessageResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Message   string      `json:"message"`
	Messages  interface{} `json:"messages"`
	Detail    string      `json:"detail"`
}
type PingResponse = DeleteResponse

type ImportResponse struct {
	Code      int             `json:"code"`
	ErrorCode string          `json:"errorCode"`
	Message   string          `json:"message"`
	Data      []service.Topic `json:"data"`
	Detail    string          `json:"detail"`
}

type GrantPermissionRequest struct {
	Actions service.Actions `json:"actions"`
}

func (c *controller) getCreateResponse(code int, errorCode string, detail string) *CreateResponse {
	return &CreateResponse{
		Code:      code,
		Message:   c.errMsg.Application[errorCode],
		Detail:    detail,
		ErrorCode: errorCode,
	}
}

func (c *controller) CreateTopic(req *restful.Request) (int, *CreateResponse) {
	tp := &service.Topic{}
	if err := req.ReadEntity(tp); err != nil {
		return http.StatusInternalServerError, c.getCreateResponse(1, tperror.Error_Create_Topic, fmt.Sprintf("cannot read entity: %+v", err))
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, c.getCreateResponse(1, tperror.Error_Auth_Error, fmt.Sprintf("auth model error: %+v", err))
	}

	tp.Users = user.InitWithOwner(authuser.Name)
	if tp, err := c.service.CreateTopic(tp); err != nil {
		return http.StatusInternalServerError, c.getCreateResponse(2, tperror.Error_Create_Topic, fmt.Sprintf("create database error: %+v", err))
	} else {
		return http.StatusOK, &CreateResponse{
			Code:   0,
			Data:   tp,
			Detail: "accepted topic create request, please waiting for create topic on pulsar",
		}
	}
}

func (c *controller) GetTopic(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if tp, err := c.service.GetTopic(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: tp,
		}
	}
}

//批量删除topics
func (c *controller) DeleteTopics(req *restful.Request) (int, *ListResponse) {
	ids := req.QueryParameters("ids")
	for _, id := range ids {
		if _, err := c.service.DeleteTopic(id); err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    1,
				Message: fmt.Errorf("delete topic error: %+v", err).Error(),
			}
		}
	}
	return http.StatusOK, &ListResponse{
		Code:    0,
		Message: "delete topic success",
	}
}
func (c *controller) DeleteTopic(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if topic, err := c.service.DeleteTopic(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete topic error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:    0,
			Data:    topic,
			Message: "deleting",
		}
	}
}

func (c *controller) ListTopic(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")

	if tp, err := c.service.ListTopic(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tps TopicList = c.ListTopicByField(req, tp)

		data, err := util.PageWrap(tps, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    1,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		} else {
			return http.StatusOK, &ListResponse{
				Code: 0,
				Data: data,
			}
		}

	}
}

//根据可选搜索字段过滤topic列表,大小写不敏感
func (c *controller) ListTopicByField(req *restful.Request, tps []*service.Topic) []*service.Topic {
	//Topic名称查询参数
	var tpsResult []*service.Topic
	name := req.QueryParameter("name")
	if len(name) == 0 {
		return tps
	} else {
		name = strings.ToLower(name)
		for _, tp := range tps {
			if strings.Contains(strings.ToLower(tp.Name), name) {
				tpsResult = append(tpsResult, tp)
			}
		}
	}
	return tpsResult
}

type TopicList []*service.Topic

func (ts TopicList) Len() int {
	return len(ts)
}

func (ts TopicList) GetItem(i int) (interface{}, error) {
	if i >= len(ts) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return ts[i], nil
}

//TODO 导入Topic待完善
func (c *controller) ImportTopics(req *restful.Request, response *restful.Response) (int, *ImportResponse) {
	spec := &parser.TopicExcelSpec{
		SheetName:        "topics",
		MultiPartFileKey: "uploadfile",
		TitleRowSpecList: []string{"topic租户名称", "topic组名称", "topic名称", "分区数量", "非持久化"},
	}
	tps, err := parser.ParseTopicsFromExcel(req, response, spec)
	if err != nil {
		return http.StatusInternalServerError, &ImportResponse{
			Code:    1,
			Message: fmt.Sprintf("import topics error: %+v", err),
			Detail:  fmt.Sprintf("import topics error: %+v", err),
		}
	}

	var data []service.Topic
	for _, tp := range tps.ParseData {

		partition, _ := strconv.Atoi(tp[3])
		isNonPersisten, _ := strconv.ParseBool(tp[4])
		topic := &service.Topic{
			Tenant:          tp[0],
			TopicGroup:      tp[1],
			Name:            tp[2],
			Partition:       partition,
			IsNonPersistent: isNonPersisten,
		}

		//TODO 数据重复判断
		//if c.service.IsTopicExist(topic) {
		//	continue
		//}

		if err := topic.Validate(); err != nil {
			return http.StatusInternalServerError, &ImportResponse{
				Code:    1,
				Message: fmt.Errorf("import topics error: %+v", err).Error(),
			}
		}
		topic.URL = topic.GetUrl()
		t, err := c.service.CreateTopic(topic)
		if err != nil {
			return http.StatusInternalServerError, &ImportResponse{
				Code:    1,
				Message: fmt.Errorf("import topics error: %+v", err).Error(),
			}
		} else {
			data = append(data, *t)
		}

	}

	return http.StatusOK, &ImportResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}

}

//查询topic的消息
func (c *controller) ListMessages(req *restful.Request) (int, *MessageResponse) {
	topicName := req.QueryParameter("topicName")
	startTime := req.QueryParameter("startTime")
	endTime := req.QueryParameter("endTime")
	topicGroup := req.QueryParameter("topicGroup")
	//先查出所有topic的信息
	tp, err := c.service.ListTopic()
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	//接收参数只有topicName
	if len(topicName) > 0 && len(topicGroup) == 0 && len(startTime) == 0 && len(endTime) == 0 {
		//通过topicName字段来匹配topic
		tps := c.ListTopicByTopicName(topicName, tp)
		var tpUrls []string
		//获取topic的url
		for _, tp := range tps {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrl(tpUrls, req)
		return httpStatus, messageResponse
	} else if len(topicGroup) > 0 && len(topicName) == 0 && len(startTime) == 0 && len(endTime) == 0 { //接收参数只有topicGroup
		//通过topicGroup字段来匹配topic
		tps := c.ListTopicByTopicGroup(topicGroup, tp)
		var tpUrls []string
		for _, tp := range tps {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrl(tpUrls, req)
		return httpStatus, messageResponse
	} else if len(startTime) > 0 && len(endTime) > 0 && len(topicGroup) == 0 && len(topicName) == 0 { //接收参数只有时间
		start, _ := strconv.ParseInt(startTime, 10, 64)
		end, _ := strconv.ParseInt(endTime, 10, 64)
		var tpUrls []string
		for _, tp := range tp {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrlTime(tpUrls, start, end, req)
		return httpStatus, messageResponse
	} else if len(topicName) > 0 && len(topicGroup) > 0 && len(startTime) == 0 && len(endTime) == 0 { //接收参数有topicName,topicGroup
		tps := c.ListTopicByTopicGroupAndName(topicGroup, topicName, tp)
		var tpUrls []string
		for _, tp := range tps {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrl(tpUrls, req)
		return httpStatus, messageResponse
	} else if len(topicGroup) > 0 && len(topicName) == 0 && len(startTime) > 0 && len(endTime) > 0 { //接收参数有topicGroup和时间
		start, _ := strconv.ParseInt(startTime, 10, 64)
		end, _ := strconv.ParseInt(endTime, 10, 64)
		tps := c.ListTopicByTopicGroup(topicGroup, tp)
		var tpUrls []string
		for _, tp := range tps {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrlTime(tpUrls, start, end, req)
		return httpStatus, messageResponse
	} else if len(topicGroup) == 0 && len(topicName) > 0 && len(startTime) > 0 && len(endTime) > 0 { //接收参数有topicName和时间
		start, _ := strconv.ParseInt(startTime, 10, 64)
		end, _ := strconv.ParseInt(endTime, 10, 64)
		tps := c.ListTopicByTopicName(topicName, tp)
		var tpUrls []string
		for _, tp := range tps {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrlTime(tpUrls, start, end, req)
		return httpStatus, messageResponse
	} else if len(topicName) > 0 && len(topicGroup) > 0 && len(startTime) > 0 && len(endTime) > 0 { //接收参数有topicName、topicGroup、时间
		start, _ := strconv.ParseInt(startTime, 10, 64)
		end, _ := strconv.ParseInt(endTime, 10, 64)
		tps := c.ListTopicByTopicGroupAndName(topicGroup, topicName, tp)
		var tpUrls []string
		for _, tp := range tps {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrlTime(tpUrls, start, end, req)
		return httpStatus, messageResponse
	} else { //没有参数
		var tpUrls []string
		for _, tp := range tp {
			tpUrls = append(tpUrls, tp.URL)
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrl(tpUrls, req)
		return httpStatus, messageResponse
	}
}

//通过topicUrl查询topic的消息(不带时间)
func (c *controller) ListMessagesByTopicUrl(topicUrls []string, req *restful.Request) (int, *MessageResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	if messages, err := c.service.ListMessages(topicUrls); err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var ms MessageList = messages
		data, err := util.PageWrap(ms, page, size)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:    1,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &MessageResponse{
			Code:     0,
			Messages: data,
		}
	}
}

//通过topicUrl查询topic的消息(带时间)
func (c *controller) ListMessagesByTopicUrlTime(topicUrls []string, start int64, end int64, req *restful.Request) (int, *MessageResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	if messages, err := c.service.ListMessagesTime(topicUrls, start, end); err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var ms MessageList = messages
		data, err := util.PageWrap(ms, page, size)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:    1,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &MessageResponse{
			Code:     0,
			Messages: data,
		}
	}
}

//通过topicName匹配topic
func (c *controller) ListTopicByTopicName(topicName string, tps []*service.Topic) []*service.Topic {
	var tpsResult []*service.Topic
	for _, tp := range tps {
		if strings.Compare(tp.Name, topicName) == 0 {
			tpsResult = append(tpsResult, tp)
		}
	}
	return tpsResult
}

//通过topicGroup匹配topic
func (c *controller) ListTopicByTopicGroup(topicGroup string, tps []*service.Topic) []*service.Topic {
	var tpsResult []*service.Topic
	for _, tp := range tps {
		if strings.Compare(tp.TopicGroup, topicGroup) == 0 {
			tpsResult = append(tpsResult, tp)
		}
	}
	return tpsResult
}

//通过topicGroup和topicName匹配topic
func (c *controller) ListTopicByTopicGroupAndName(topicGroup string, topicName string, tps []*service.Topic) []*service.Topic {
	var tpsResult []*service.Topic
	for _, tp := range tps {
		if strings.Compare(tp.TopicGroup, topicGroup) == 0 && strings.Compare(tp.Name, topicName) == 0 {
			tpsResult = append(tpsResult, tp)
		}
	}
	return tpsResult
}

type MessageList []service.Message

func (ms MessageList) Len() int {
	return len(ms)
}

func (ms MessageList) GetItem(i int) (interface{}, error) {
	if i >= len(ms) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return ms[i], nil
}

//导出topics的信息
func (c *controller) ExportTopics(req *restful.Request) {
	topicIds := req.QueryParameters("topicIds")
	file := excelize.NewFile()
	index := file.NewSheet("topics")
	s := []string{"topic租户名称", "topic组名称", "topic名称", "分区数量", "非持久化"}
	j := 0
	for i := 65; i < 70; i++ {
		file.SetCellValue("topics", string(i)+"1", s[j])
		j++
	}
	row := 1
	for _, topicId := range topicIds {
		row++
		cell := 65
		if topic, err := c.service.GetTopic(topicId); err != nil {
			log.Printf("list database error: %+v", err)
		} else {
			//以坐标位置写入
			file.SetCellValue("topics", string(cell)+strconv.Itoa(row), topic.Tenant)
			file.SetCellValue("topics", string(cell+1)+strconv.Itoa(row), topic.TopicGroup)
			file.SetCellValue("topics", string(cell+2)+strconv.Itoa(row), topic.Name)
			file.SetCellValue("topics", string(cell+3)+strconv.Itoa(row), topic.Partition)
			file.SetCellValue("topics", string(cell+4)+strconv.Itoa(row), topic.IsNonPersistent)
		}
	}
	file.SetActiveSheet(index)
	err := file.SaveAs("/tmp/topics.xlsx")
	if err != nil {
		log.Printf("save file error: %+v", err)
	}
}

func (c *controller) GrantPermissions(req *restful.Request) (int, *GrantResponse) {
	id := req.PathParameter("id")
	authUserId := req.PathParameter("auth-user-id")

	actions := &GrantPermissionRequest{}
	if err := req.ReadEntity(actions); err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:    1,
			Message: fmt.Sprintf("grant permissions error: %+v", err),
		}
	}

	if tp, err := c.service.GrantPermissions(id, authUserId, actions.Actions); err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GrantResponse{
			Code: 0,
			Data: tp,
		}
	}

}
func (c *controller) DeletePermissions(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authUserId := req.PathParameter("auth-user-id")
	if topic, err := c.service.DeletePermissions(id, authUserId); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete permissions error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:    0,
			Data:    topic,
			Message: "deleting",
		}
	}
}

type PermissionList []service.Permission

func (pers PermissionList) Len() int {
	return len(pers)
}

func (pers PermissionList) GetItem(i int) (interface{}, error) {
	if i >= len(pers) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return pers[i], nil
}
func (c *controller) ListUsers(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	topicId := req.PathParameter("id")
	AuthUserName := req.QueryParameter("name")
	tp, err := c.service.GetTopic(topicId)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	var permissionList PermissionList = tp.Permissions
	if len(AuthUserName) > 0 {
		permissionList = c.ListUsersByName(AuthUserName, permissionList)
	}
	data, err := util.PageWrap(permissionList, page, size)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Sprintf("page parameter error: %+v", err),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: data,
		}
	}
}

//根据用户名查询topic授权用户，模糊查询
func (c *controller) ListUsersByName(name string, per []service.Permission) []service.Permission {
	var permissionsResult []service.Permission
	name = strings.ToLower(name)
	for _, p := range per {
		if strings.Contains(strings.ToLower(p.AuthUserName), name) {
			permissionsResult = append(permissionsResult, p)
		}
	}
	return permissionsResult
}
func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
