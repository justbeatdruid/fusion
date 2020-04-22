package topic

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	tperror "github.com/chinamobile/nlpt/apiserver/resources/topic/error"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/parser"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	tgerror "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/error"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/chinamobile/nlpt/pkg/go-restful/log"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TopicConfig, cfg.LocalConfig),
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

type StatisticsResponse = struct {
	Code      int                 `json:"code"`
	ErrorCode string              `json:"errorCode"`
	Message   string              `json:"message"`
	Data      *service.Statistics `json:"data"`
	Detail    string              `json:"detail"`
}
type PingResponse = DeleteResponse

type ImportResponse struct {
	Code      int             `json:"code"`
	ErrorCode string          `json:"errorCode"`
	Message   string          `json:"message"`
	Data      []service.Topic `json:"data"`
	Detail    string          `json:"detail"`
}

type TopicSlice TopicList
type MessageSlice MessageList

const (
	success = iota
	fail
)

type GrantPermissionRequest struct {
	Actions service.Actions `json:"actions"`
}

//重写Interface的len方法
func (t TopicSlice) Len() int {
	return len(t)
}

//重写Interface的Swap方法
func (t TopicSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

//重写Interface的Less方法
func (t TopicSlice) Less(i, j int) bool {
	return t[i].CreatedAt > t[j].CreatedAt
}

func (m MessageSlice) Len() int {
	return len(m)
}

func (m MessageSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m MessageSlice) Less(i, j int) bool {
	return m[i].Time.Unix() > m[j].Time.Unix()
}

func (c *controller) getCreateResponse(code int, errorCode string, detail string, msg string) *CreateResponse {
	resp := &CreateResponse{
		Code:      code,
		Detail:    detail,
		ErrorCode: errorCode,
		Message:   msg,
	}

	if len(msg) == 0 {
		resp.Message = c.errMsg.Topic[resp.ErrorCode]
	}
	return resp
}

func (c *controller) CreateTopic(req *restful.Request) (int, *CreateResponse) {
	tp := &service.Topic{}
	if err := req.ReadEntity(tp); err != nil {
		return http.StatusInternalServerError, c.getCreateResponse(fail, tperror.ErrorReadEntity, fmt.Sprintf("cannot read entity: %+v", err), "")
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, c.getCreateResponse(fail, tperror.ErrorAuthError, fmt.Sprintf("auth model error: %+v", err), "")
	}

	tp.Users = user.InitWithOwner(authuser.Name)
	tp.Namespace = authuser.Namespace
	if tp, tpErr := c.service.CreateTopic(tp); tpErr.Err != nil {
		return http.StatusInternalServerError, c.getCreateResponse(2, tpErr.ErrorCode, fmt.Sprintf("create topic error: %+v", tpErr.Err), "")
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      success,
			ErrorCode: tperror.Success,
			Data:      tp,
			Detail:    "accepted topic create request, please waiting for create topic on pulsar",
		}
	}
}

func (c *controller) GetTopic(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
	if tp, err := c.service.GetTopic(id, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorGetTopicInfo,
			Message:   c.errMsg.Topic[tperror.ErrorGetTopicInfo],
			Detail:    fmt.Sprintf("get database error: %+v", err),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      success,
			ErrorCode: tperror.Success,
			Data:      tp,
			Detail:    "get topiclist sucecssfully",
		}
	}
}

func (c *controller) DoStatisticsOnTopics(req *restful.Request) (int, *StatisticsResponse) {
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Detail:    "",
		}
	}
	tp, err := c.service.ListTopic(util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      fail,
			ErrorCode: "",
			Message:   "",
			Detail:    fmt.Sprintf("do statistics on topics error, %+v", err),
		}

	}
	var data = c.CountTopics(tp)
	return http.StatusOK, &StatisticsResponse{
		Code:      success,
		ErrorCode: tgerror.Success,
		Data:      data,
		Detail:    "do statistics on topics successfully",
	}
}

func (c *controller) CountTopics(tp []*service.Topic) *service.Statistics {
	data := &service.Statistics{}
	data.Total = len(tp)

	var totalMessageSize int64

	var increment int
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	for _, t := range tp {
		if t.CreatedAt < end.Unix() && t.CreatedAt >= start.Unix() {
			increment++
		}
		totalMessageSize += t.Stats.BytesInCounter

	}
	data.Increment = increment
	data.MessageSize = c.formatSize(totalMessageSize)
	return data
}

// 字节的单位转换 保留两位小数
func (c *controller) formatSize(messageSize int64) (size string) {
	if messageSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.fB", float64(messageSize)/float64(1))
	} else if messageSize < (1024 * 1024) {
		return fmt.Sprintf("%.fKB", float64(messageSize)/float64(1024))
	} else if messageSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.fMB", float64(messageSize)/float64(1024*1024))
	} else if messageSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.fGB", float64(messageSize)/float64(1024*1024*1024))
	} else if messageSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.fTB", float64(messageSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.fEB", float64(messageSize)/float64(1024*1024*1024*1024*1024))
	}
}

//批量删除topics
/*func (c *controller) DeleteTopics(req *restful.Request) (int, *ListResponse) {
	ids := req.QueryParameters("ids")
	for _, id := range ids {
		if _, err := c.service.DeleteTopic(id); err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("delete topic error: %+v", err).Error(),
			}
		}
	}
	return http.StatusOK, &ListResponse{
		Code:    success,
		Message: "delete topic success",
	}
}*/
func (c *controller) DeleteTopic(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
	if topic, err := c.service.DeleteTopic(id, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorDeleteTopic,
			Message:   c.errMsg.Topic[tperror.ErrorDeleteTopic],
			Detail:    fmt.Sprintf("delete topic error: %+v", err),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      success,
			ErrorCode: tperror.Success,
			Data:      topic,
			Message:   "deleting",
		}
	}
}

func (c *controller) ListTopic(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
	if tp, err := c.service.ListTopic(util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorGetTopicList,
			Message:   c.errMsg.Topic[tperror.ErrorGetTopicList],
			Detail:    fmt.Sprintf("list database error: %+v", err),
		}
	} else {
		var tps TopicList = c.ListTopicByField(req, tp)
		//按照创建时间排序
		sort.Sort(TopicSlice(tps))
		data, err := util.PageWrap(tps, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorPageParamInvalid,
				Message:   c.errMsg.Topic[tperror.ErrorPageParamInvalid],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		} else {
			return http.StatusOK, &ListResponse{
				Code:      success,
				ErrorCode: tperror.Success,
				Data:      data,
				Detail:    "list topic successfully",
			}
		}

	}
}

//根据可选搜索字段过滤topic列表,大小写不敏感
func (c *controller) ListTopicByField(req *restful.Request, tps []*service.Topic) []*service.Topic {
	name := req.QueryParameter("name")
	topicGroup := req.QueryParameter("topicGroup")
	//Topic名称查询参数
	if len(name) > 0 {
		tps = c.ListTopicByTopicName(name, tps)
	}
	//TopicGroup查询参数
	if len(topicGroup) > 0 {
		tps = c.ListTopicByTopicGroup(topicGroup, tps)
	}

	return tps
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
			Code:      1,
			ErrorCode: tperror.ErrorParseImportFile,
			Message:   c.errMsg.Topic[tperror.ErrorParseImportFile],
			Detail:    fmt.Sprintf("import topics error: %+v", err),
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

		if tpErr := topic.Validate(); tpErr.Err != nil {
			return http.StatusInternalServerError, &ImportResponse{
				Code:      1,
				ErrorCode: tperror.ErrorImportBadRequest,
				Message:   c.errMsg.Topic[tperror.ErrorImportBadRequest],
				Detail:    fmt.Sprintf("import topics error: %+v", err),
			}
		}
		topic.URL = topic.GetUrl()
		authuser, err := auth.GetAuthUser(req)
		if err != nil {
			return http.StatusInternalServerError, &ImportResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorAuthError,
				Message:   fmt.Sprintf("auth model error: %+v", err),
				Data:      nil,
				Detail:    "",
			}
		}
		//TODO 数据重复判断
		if c.service.IsTopicUrlExist(topic.GetUrl()) {
			return http.StatusInternalServerError, &ImportResponse{
				Code:      1,
				ErrorCode: tperror.ErrorTopicExists,
				Message:   c.errMsg.Topic[tperror.ErrorTopicExists],
				Detail:    fmt.Sprintf("import topics error: %+v", err),
			}
		}
		topic.Users = user.InitWithOwner(authuser.Name)
		topic.Namespace = authuser.Namespace
		// 创建Topic
		t, tpErr := c.service.CreateTopic(topic)
		if tpErr.Err != nil {
			return http.StatusInternalServerError, &ImportResponse{
				Code:      1,
				ErrorCode: tpErr.ErrorCode,
				Message:   c.errMsg.Topic[tpErr.ErrorCode],
				Detail:    fmt.Sprintf("import topics error: %+v", err),
			}
		} else {
			data = append(data, *t)
		}

	}

	return http.StatusOK, &ImportResponse{
		Code:    success,
		Message: tperror.Success,
		Data:    data,
	}

}

//查询topic的消息
func (c *controller) ListMessages(req *restful.Request) (int, *MessageResponse) {
	topicName := req.QueryParameter("topicName")
	startTime := req.QueryParameter("startTime")
	endTime := req.QueryParameter("endTime")
	topicGroup := req.QueryParameter("topicGroup")

	topicName = "testTopic"
	topicGroup = "default"
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Messages:  nil,
			Detail:    "",
		}
	}
	authUser.Namespace = "public"
	//先查出所有topic的信息
	tps, err := c.service.ListTopic(util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorGetTopicList]),
			Detail:    fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	//topicName筛选
	if len(topicName) > 0 {
		//通过topicName字段来匹配topic
		tps = c.ListTopicByTopicName(topicName, tps)
	}
	//topicGroup筛选
	if len(topicGroup) > 0 {
		//通过topicGroup字段来匹配topic
		tps = c.ListTopicByTopicGroup(topicGroup, tps)
	}
	var tpUrls []string
	for _, tp := range tps {
		tpUrls = append(tpUrls, tp.URL)
	}
	//时间筛选
	if len(startTime) > 0 && len(endTime) > 0 {
		start, err := strconv.ParseInt(startTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorQueryMessageStartTime,
				Message:   c.errMsg.Topic[tperror.ErrorQueryMessageStartTime],
				Detail:    fmt.Sprintf("startTime parameter error: %+v", err),
			}
		}
		end, err := strconv.ParseInt(endTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorQueryMessageEndTime,
				Message:   c.errMsg.Topic[tperror.ErrorQueryMessageEndTime],
				Detail:    fmt.Sprintf("endTime parameter error: %+v", err),
			}
		}
		httpStatus, messageResponse := c.ListMessagesByTopicUrlTime(tpUrls, start, end, req)
		return httpStatus, messageResponse
	} else { //没有时间
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
			Code:    fail,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var ms MessageList = messages
		sort.Sort(MessageSlice(ms))
		data, err := util.PageWrap(ms, page, size)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorQueryMessagePageParam,
				Message:   c.errMsg.Topic[tperror.ErrorQueryMessagePageParam],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &MessageResponse{
			Code:     success,
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
			Code:      fail,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], err.Error()),
			Detail:    fmt.Sprintf("list database error: %+v", err),
		}
	} else {
		var ms MessageList = messages
		sort.Sort(MessageSlice(ms))
		data, err := util.PageWrap(ms, page, size)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorQueryMessagePageParam,
				Message:   c.errMsg.Topic[tperror.ErrorQueryMessagePageParam],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &MessageResponse{
			Code:      success,
			ErrorCode: tperror.Success,
			Messages:  data,
		}
	}
}

//通过topicName匹配topic(模糊匹配)
func (c *controller) ListTopicByTopicName(topicName string, tps []*service.Topic) []*service.Topic {
	var tpsResult []*service.Topic
	topicName = strings.ToLower(topicName)
	for _, tp := range tps {
		if strings.Contains(strings.ToLower(tp.Name), topicName) {
			tpsResult = append(tpsResult, tp)
		}
	}
	return tpsResult
}

//通过topicGroup匹配topic(模糊匹配)
func (c *controller) ListTopicByTopicGroup(topicGroup string, tps []*service.Topic) []*service.Topic {
	var tpsResult []*service.Topic
	topicGroup = strings.ToLower(topicGroup)
	for _, tp := range tps {
		if strings.Contains(strings.ToLower(tp.TopicGroup), topicGroup) {
			tpsResult = append(tpsResult, tp)
		}
	}
	return tpsResult
}

/*//通过topicGroup和topicName匹配topic
func (c *controller) ListTopicByTopicGroupAndName(topicGroup string, topicName string, tps []*service.Topic) []*service.Topic {
	var tpsResult []*service.Topic
	for _, tp := range tps {
		if strings.Compare(tp.TopicGroup, topicGroup) == 0 && strings.Compare(tp.Name, topicName) == 0 {
			tpsResult = append(tpsResult, tp)
		}
	}
	return tpsResult
}*/

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
			Code:      fail,
			ErrorCode: tperror.ErrorReadEntity,
			Message:   c.errMsg.Topic[tperror.ErrorReadEntity],
			Detail:    fmt.Sprintf("grant permissions error: %+v", err),
		}
	}
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
	if tp, err := c.service.GrantPermissions(id, authUserId, actions.Actions, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:      2,
			ErrorCode: tperror.ErrorGrantPermissions,
			Message:   c.errMsg.Topic[tperror.ErrorGrantPermissions],
			Detail:    fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GrantResponse{
			Code:      success,
			ErrorCode: tperror.Success,
			Data:      tp,
		}
	}

}
func (c *controller) DeletePermissions(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authUserId := req.PathParameter("auth-user-id")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
	if topic, err := c.service.DeletePermissions(id, authUserId, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorRevokePermissions,
			Message:   c.errMsg.Topic[tperror.ErrorRevokePermissions],
			Detail:    fmt.Sprintf("delete permissions error: %+v", err),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      success,
			Data:      topic,
			ErrorCode: tperror.Success,
			Detail:    "accepted topic delete request",
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
