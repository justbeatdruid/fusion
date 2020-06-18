package topic

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	tperror "github.com/chinamobile/nlpt/apiserver/resources/topic/error"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/parser"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/pulsarsql"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	tgerror "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/error"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	v1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type controller struct {
	service  *service.Service
	errMsg   config.ErrorConfig
	tpConfig *config.TopicConfig
}

const (
	//order        = `select * from (select row_number() over(order by __publish_time__ desc) __row__, * from pulsar."%s/%s"."%s" %s) as t where __row__ between %d and %d`
	messageIdSql = `WHERE "__message_id__" = '%s' AND "__partition__" = %s`
	keySql       = `WHERE "__key__" LIKE '%%%s%%'`
	timeSql      = `WHERE "__publish_time__" BETWEEN timestamp '%s' AND timestamp '%s'`
	order        = `WITH subquery_1 AS (
    SELECT count("__message_id__") as __count__ from pulsar."%s/%s"."%s" %s),
subquery_2 AS (
   select * from (select row_number() over(order by __publish_time__ desc) row, * from pulsar."%s/%s"."%s" %s) as t where row between %d and %d
)   
select * From subquery_1, subquery_2`
)

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TopicConfig, cfg.LocalConfig),
		cfg.LocalConfig,
		cfg.TopicConfig,
	}
}

type Wrapped struct {
	Code      int            `json:"code"`
	ErrorCode string         `json:"errorCode"`
	Message   string         `json:"message"`
	Data      *service.Topic `json:"data,omitempty"`
	Detail    string         `json:"detail"`
}

type RequestWrapped struct {
	Data *service.Topic `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = RequestWrapped
type DeleteResponse = Wrapped
type GrantResponse = Wrapped
type ExportResponse = Wrapped
type ResetPositionResponse = Wrapped
type AddPartitions = Wrapped
type SendMessagesResponse = ListResponse
type BatchGrantResponse = ListResponse

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
	Data      interface{} `json:"data"`
	Detail    string      `json:"detail"`
	TotalSize interface{} `json:"totalSize"`
	Page      int         `json:"page"`
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

type BindOrReleaseRequest struct {
	Operation string             `json:"operation"`
	Topics    []service.BindInfo `json:"topics"`
}
type BindOrReleaseResponse struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
}

type SubscriptionsResponse struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Detail    string      `json:"detail"`
}
type TopicSlice TopicList
type MessageSlice MessageList

const (
	success = iota
	fail
)

type GrantPermissionRequest struct {
	Actions v1.Actions `json:"actions"`
}
type BatchGrantPermissionRequest struct {
	ClientAuths []service.GrantPermissions `json:"clientAuths"`
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
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
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
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
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
func (c *controller) DeleteTopics(req *restful.Request) (int, *ListResponse) {
	ids := req.QueryParameters("ids")
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
	for _, id := range ids {
		if _, err := c.service.DeleteTopic(id, util.WithNamespace(authUser.Namespace)); err != nil {
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
}
func (c *controller) DeleteTopic(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
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
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
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
	application := req.QueryParameter("application")
	//Topic名称查询参数
	if len(name) > 0 {
		tps = c.ListTopicByTopicName(name, tps)
	}
	//TopicGroup查询参数
	if len(topicGroup) > 0 {
		tps = c.ListTopicByTopicGroup(topicGroup, tps)
	}

	if len(application) > 0 {
		tps = c.ListTopicByApplication(application, tps)
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
		TitleRowSpecList: []string{"topic租户名称", "topic组名称", "topic名称", "多分区", "分区数量", "持久化"},
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
	//var tgmap = make(map[string]string)

	for _, tp := range tps.ParseData {

		partitioned, _ := strconv.ParseBool(tp[3])
		partitionNum, _ := strconv.Atoi(tp[4])
		persistent, _ := strconv.ParseBool(tp[5])

		//根据topicGroup名称查找当前租户下是否有同名的资源

		topic := &service.Topic{
			//Tenant:       tp[0],
			TopicGroup:   tp[1],
			Name:         tp[2],
			Partitioned:  &partitioned,
			PartitionNum: partitionNum,
			Persistent:   &persistent,
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
		if c.service.IsTopicUrlExist(topic.GetUrl(), util.WithNamespace(authuser.Namespace)) {
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

	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
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
			Code: success,
			Data: data,
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
			Data:      data,
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

func (c *controller) ListTopicByApplication(application string, tps []*service.Topic) []*service.Topic {
	var tpsResult []*service.Topic

	for _, tp := range tps {
		for _, app := range tp.Applications {
			if app.ID == application {
				tpsResult = append(tpsResult, tp)
				continue
			}
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
type MessagesList []service.Messages

func (ms MessageList) Len() int {
	return len(ms)
}

func (ms MessageList) GetItem(i int) (interface{}, error) {
	if i >= len(ms) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return ms[i], nil
}

func (msl MessagesList) Len() int {
	return len(msl)
}

func (msl MessagesList) GetItem(i int) (interface{}, error) {
	if i >= len(msl) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return msl[i], nil
}

//导出topics的信息
func (c *controller) ExportTopics(req *restful.Request) (int, *ExportResponse) {
	topicIds := req.QueryParameters("topicIds")
	file := excelize.NewFile()
	index := file.NewSheet("topics")
	s := []string{"topic租户名称", "topic组名称", "topic名称", "多分区", "分区数量", "持久化"}
	j := 0
	for i := 65; i <= 70; i++ {
		file.SetCellValue("topics", string(i)+"1", s[j])
		j++
	}
	row := 1
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ExportResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("auth model error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
	for _, topicId := range topicIds {
		row++
		cell := 65
		if topic, err := c.service.GetTopic(topicId, util.WithNamespace(authUser.Namespace)); err != nil {
			return http.StatusInternalServerError, &ExportResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorAuthError,
				Message:   c.errMsg.Topic[tperror.ErrorAuthError],
				Data:      nil,
				Detail:    fmt.Sprintf("auth model error: %+v", err),
			}
		} else {
			//以坐标位置写入
			file.SetCellValue("topics", string(cell)+strconv.Itoa(row), topic.Namespace)
			file.SetCellValue("topics", string(cell+1)+strconv.Itoa(row), topic.TopicGroup)
			file.SetCellValue("topics", string(cell+2)+strconv.Itoa(row), topic.Name)
			file.SetCellValue("topics", string(cell+3)+strconv.Itoa(row), *topic.Partitioned)
			file.SetCellValue("topics", string(cell+4)+strconv.Itoa(row), topic.PartitionNum)
			file.SetCellValue("topics", string(cell+5)+strconv.Itoa(row), *topic.Persistent)
		}
	}
	file.SetActiveSheet(index)
	err = file.SaveAs("/tmp/topics.xlsx")
	if err != nil {
		return http.StatusInternalServerError, &ExportResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   fmt.Sprintf("save file error: %+v", err),
			Data:      nil,
			Detail:    "",
		}
	}
	return http.StatusOK, &ExportResponse{
		Code:    success,
		Message: tperror.Success,
		Data:    nil,
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
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	if tp, err := c.service.GrantPermissions(id, authUserId, actions.Actions, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:      2,
			ErrorCode: tperror.ErrorGrantPermissions,
			Message:   c.errMsg.Topic[tperror.ErrorGrantPermissions],
			Detail:    fmt.Errorf("grant permissions error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GrantResponse{
			Code:      success,
			ErrorCode: tperror.Success,
			Data:      tp,
		}
	}

}

func (c *controller) ModifyPermissions(req *restful.Request) (int, *GrantResponse) {
	id := req.PathParameter("id")
	authUserId := req.PathParameter("auth-user-id")

	actions := &GrantPermissionRequest{}
	if err := req.ReadEntity(actions); err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorReadEntity,
			Message:   c.errMsg.Topic[tperror.ErrorReadEntity],
			Detail:    fmt.Sprintf("modify permissions error: %+v", err),
		}
	}
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	if tp, err := c.service.ModifyPermissions(id, authUserId, actions.Actions, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GrantResponse{
			Code:      2,
			ErrorCode: tperror.ErrorModifyPermissions,
			Message:   c.errMsg.Topic[tperror.ErrorModifyPermissions],
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
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
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
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	tp, err := c.service.GetTopic(topicId, util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	}

	var permissionList PermissionList = tp.Permissions

	var permissions = make([]service.Permission, 0)
	for _, p := range permissionList {
		ca, err := c.service.QueryAuthUserById(p.AuthUserID, util.WithNamespace(authUser.Namespace))
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    1,
				Message: fmt.Errorf("list database error: %+v", err).Error(),
			}
		}
		p.Token = ca.Spec.Token
		p.IssuedAt = ca.Spec.IssuedAt
		p.ExpireAt = ca.Spec.ExipreAt
		if ca.Spec.ExipreAt > util.Now().Unix() {
			p.Effective = true
		} else {
			p.Effective = false
		}

		permissions = append(permissions, p)
	}
	permissionList = permissions
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

func (c *controller) QueryMessage(req *restful.Request) (int, *MessageResponse) {
	messageId := req.QueryParameter("messageId")
	topic := req.QueryParameter("topic")
	topicGroup := req.QueryParameter("topicGroup")
	key := req.QueryParameter("key")
	startTime := req.QueryParameter("startTime")
	endTime := req.QueryParameter("endTime")
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	//参数判断,三个有且只有一个有值
	httpStatus, messageResponse := c.QueryParamterCheck(messageId, key, startTime, endTime)
	if messageResponse != nil {
		return httpStatus, messageResponse
	}
	var sql string
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	//判断topic和topicGroup是否存在
	h, m, topic := c.TopicCheck(topic, authUser)
	if m != nil {
		return h, m
	}
	h, m, topicGroup = c.TopicGroupCheck(topicGroup, authUser)
	if m != nil {
		return h, m
	}
	tenant := authUser.Namespace
	//分页
	if len(page) == 0 {
		page = "1"
	}
	if len(size) == 0 {
		size = "10"
	}
	p, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   c.errMsg.Topic[tperror.ErrorQueryMessage],
			Detail:    fmt.Sprintf("page error: %+v ", err),
		}
	}
	s, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   c.errMsg.Topic[tperror.ErrorQueryMessage],
			Detail:    fmt.Errorf("size error: %+v", err).Error(),
		}
	}
	start := (p-1)*s + 1
	end := p * s
	//根据MessageID查询topic信息
	if len(messageId) > 0 {
		//通过pulsar客户端查出来的id格式为622:1:-1:0
		//将冒号换成逗号
		messageId = strings.ReplaceAll(messageId, ":", ",")
		//两边没有括号就加上括号
		h, m, ok := c.QueryParamterValidate(`^\(`, messageId)
		if ok == 1 {
			return h, m
		} else if ok == 2 {
			messageId = "(" + messageId + ")"
		}
		//reg := `^\([0-9]{1,}\,[0-9]{1,}\,-1|[0-9]{1,},[0-9]{1,}\)$`
		//校验messageId的合法性
		//h, m, ok = c.QueryParamterValidate(reg, messageId)
		//if ok != 0 {
		//	return h, m
		//}

		/*第三个数字表示分区*/
		idSplit := strings.Split(messageId, ",")
		partition := idSplit[2]
		messageId = strings.Join(append(idSplit[0:2], idSplit[3]), ",")
		sql = fmt.Sprintf(messageIdSql, messageId, partition)
		sql = fmt.Sprintf(order, tenant, topicGroup, topic, sql, tenant, topicGroup, topic, sql, start, end)
	}
	//根据key查询topic信息
	if len(key) > 0 {
		sql = fmt.Sprintf(keySql, key)
		sql = fmt.Sprintf(order, tenant, topicGroup, topic, sql, tenant, topicGroup, topic, sql, start, end)
	}
	//根据time查询topic信息
	if len(startTime) > 0 && len(endTime) > 0 {
		//校验时间的合法性
		//reg := `/^[1-9]\d{3}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])\s+(20|21|22|23|[0-1]\d):[0-5]\d:[0-5]\d$/`
		//h, m, ok := c.QueryParamterValidate(reg, startTime)
		//if ok != 0 {
		//	return h, m
		//}
		//h, m, ok = c.QueryParamterValidate(reg, endTime)
		//if ok != 0 {
		//	return h, m
		//}
		startTimeInt64, err := strconv.ParseInt(startTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      1,
				ErrorCode: tperror.ErrorQueryMessage,
				Message:   c.errMsg.Topic[tperror.ErrorQueryMessage],
				Detail:    fmt.Errorf("startTime error: %+v", err).Error(),
			}
		}

		endTimeInt64, err := strconv.ParseInt(endTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      1,
				ErrorCode: tperror.ErrorQueryMessage,
				Message:   c.errMsg.Topic[tperror.ErrorQueryMessage],
				Detail:    fmt.Errorf("endTime error: %+v", err).Error(),
			}
		}

		startTime = time.Unix(startTimeInt64, 0).Format("2006-01-02 15:04:05")
		endTime = time.Unix(endTimeInt64, 0).Format("2006-01-02 15:04:05")
		sql = fmt.Sprintf(timeSql, startTime, endTime)
		sql = fmt.Sprintf(order, tenant, topicGroup, topic, sql, tenant, topicGroup, topic, sql, start, end)
	}
	httpStatus, messageResponse = c.QueryTopicMessage(sql)
	messageResponse.Page = int(p)
	return httpStatus, messageResponse
}
func (c *controller) QueryTopicMessage(sql string) (int, *MessageResponse) {
	connector := pulsarsql.Connector{
		PrestoUser: "test-user",
		Host:       c.tpConfig.PrestoHost,
		Port:       c.tpConfig.PrestoPort,
	}
	messages, err := pulsarsql.QueryTopicMessages(connector, sql)
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   c.errMsg.Topic[tperror.ErrorQueryMessage],
			Detail:    fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	if messages != nil {
		return http.StatusOK, &MessageResponse{
			Code:      success,
			Data:      messages,
			TotalSize: messages[0].Total,
		}
	} else {
		return http.StatusOK, &MessageResponse{
			Code:      success,
			Data:      messages,
			TotalSize: 0,
		}
	}

}

//查询参数是否有值判断
func (c *controller) QueryParamterCheck(messageId string, key string, startTime string, endTime string) (int, *MessageResponse) {
	if !((len(messageId) > 0 && len(key) == 0 && len(startTime) == 0 && len(endTime) == 0) || (len(messageId) == 0 && len(key) > 0 && len(startTime) == 0 && len(endTime) == 0) || (len(messageId) == 0 && len(key) == 0 && len(startTime) > 0 && len(endTime) > 0)) {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorQueryParameterError]),
			Detail:    fmt.Errorf("query parameter error: ").Error(),
		}
	} else {
		return http.StatusOK, nil
	}
}

//判断查询参数topic和topicGroup是否存在
func (c *controller) TopicCheck(topic string, authUser auth.AuthUser) (int, *MessageResponse, string) {
	//判断topic是否存在
	if len(topic) > 0 {
		t, err := c.service.GetTopic(topic, util.WithNamespace(authUser.Namespace))
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      1,
				ErrorCode: tperror.ErrorQueryMessage,
				Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorGetTopicInfo]),
				Detail:    fmt.Errorf("list database error: %+v", err).Error(),
			}, ""
		}
		if t == nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      1,
				ErrorCode: tperror.ErrorQueryMessage,
				Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorCannotFindTopic]),
				Detail:    fmt.Errorf("topic is not exist: %+v", err).Error(),
			}, ""
		}
		topic = t.Name
		return http.StatusOK, nil, topic
	} else {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorTopicIdError]),
			Detail:    fmt.Errorf("topic paramter error ").Error(),
		}, ""
	}
}

//判断topicGroup是否存在
func (c *controller) TopicGroupCheck(topicGroup string, authUser auth.AuthUser) (int, *MessageResponse, string) {
	if len(topicGroup) > 0 {
		tg, err := c.service.GetTopicgroup(topicGroup, authUser.Namespace)
		if err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      1,
				ErrorCode: tperror.ErrorQueryMessage,
				Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorGetTopicGroupInfo]),
				Detail:    fmt.Errorf("list database error: %+v", err).Error(),
			}, ""
		}
		if tg == nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:      1,
				ErrorCode: tperror.ErrorQueryMessage,
				Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorCannotFindTopicgroup]),
				Detail:    fmt.Errorf("list database error: %+v", err).Error(),
			}, ""
		}
		topicGroup = tg.Spec.Name
		return http.StatusOK, nil, topicGroup
	} else {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], c.errMsg.Topic[tperror.ErrorTopicGroupIdError]),
			Detail:    fmt.Errorf("list database error ").Error(),
		}, ""
	}
}

//校验参数的合法性
func (c *controller) QueryParamterValidate(reg string, param string) (int, *MessageResponse, int) {
	r, err := regexp.Compile(reg)
	if err != nil {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   c.errMsg.Topic[tperror.ErrorQueryMessage],
			Detail:    fmt.Errorf("regexp error: %+v", err).Error(),
		}, 1
	}
	ok := r.MatchString(param)
	if !ok {
		return http.StatusInternalServerError, &MessageResponse{
			Code:      1,
			ErrorCode: tperror.ErrorQueryMessage,
			Message:   fmt.Sprintf(c.errMsg.Topic[tperror.ErrorQueryMessage], tperror.ErrorQueryParameterError),
			Detail:    fmt.Errorf("messageId is not legal: %+v", err).Error(),
		}, 2
	} else {
		return http.StatusOK, nil, 0
	}
}

func (c *controller) AddPartitionsOfTopic(req *restful.Request) (int, *AddPartitions) {
	id := req.PathParameter("id")
	partitions := req.PathParameter("partitions")
	partitionNum, _ := strconv.ParseInt(partitions, 10, 8)
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &AddPartitions{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	topic, err := c.service.AddPartitionsOfTopic(id, int(partitionNum), util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &AddPartitions{
			Code:      fail,
			ErrorCode: tperror.ErrorAddPartitionsOfTopicError,
			Message:   c.errMsg.Topic[tperror.ErrorAddPartitionsOfTopicError],
			Data:      nil,
			Detail:    fmt.Sprintf("add partition fail: %+v ", err),
		}
	} else {
		return http.StatusOK, &AddPartitions{
			Code: 0,
			Data: topic,
		}
	}

}

func (c *controller) BatchBindOrReleaseApi(req *restful.Request) (int, *BindOrReleaseResponse) {
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &BindOrReleaseResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}

	body := &BindOrReleaseRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindOrReleaseResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorReadEntity,
			Message:   c.errMsg.Topic[tperror.ErrorReadEntity],
			Detail:    fmt.Sprintf("%+v topics error: %+v ", body.Operation, err),
		}
	}

	appid := req.PathParameter("app-id")
	if len(appid) == 0 {
		return http.StatusInternalServerError, &BindOrReleaseResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorReadEntity,
			Message:   c.errMsg.Topic[tperror.ErrorReadEntity],
			Detail:    fmt.Sprintf("%+v topics error: %+v", body.Operation, err),
		}
	}

	if err := c.service.BatchBindOrRelease(appid, body.Operation, body.Topics, util.WithNamespace(authUser.Namespace)); err != nil {

		var errorCode string
		if body.Operation == "bind" {
			errorCode = tperror.ErrorBindTopicError
		} else {
			errorCode = tperror.ErrorUnBindTopicError
		}
		return http.StatusInternalServerError, &BindOrReleaseResponse{
			Code:      fail,
			ErrorCode: errorCode,
			Message:   c.errMsg.Topic[errorCode],
			Detail:    fmt.Sprintf("%+v topics error: %+v", body.Operation, err),
		}
	}

	return http.StatusOK, &BindOrReleaseResponse{
		Code:      success,
		ErrorCode: tperror.Success,
		Detail:    "success",
	}
}

func (c *controller) GetSubscriptionsOfTopic(req *restful.Request) (int, *SubscriptionsResponse) {
	id := req.PathParameter("id")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &SubscriptionsResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}

	t, err := c.service.GetTopic(id, util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &SubscriptionsResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorQuerySubscriptionsInfo,
			Message:   c.errMsg.Topic[tperror.ErrorQuerySubscriptionsInfo],
			Data:      nil,
			Detail:    fmt.Sprintf("query subscription error: %+v", err),
		}
	}

	return http.StatusOK, &SubscriptionsResponse{
		Code:      success,
		ErrorCode: tperror.Success,
		Data:      c.service.GetSubscriptionsOfTopic(t),
		Detail:    fmt.Sprintf("query subscription success"),
	}
}

func (c *controller) GetPartitionedSubscritionsOfTopic(req *restful.Request) (int, *SubscriptionsResponse) {
	id := req.PathParameter("id")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &SubscriptionsResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}

	t, err := c.service.GetTopic(id, util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &SubscriptionsResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorQuerySubscriptionsInfo,
			Message:   c.errMsg.Topic[tperror.ErrorQuerySubscriptionsInfo],
			Data:      nil,
			Detail:    fmt.Sprintf("query subscription error: %+v", err),
		}
	}

	if *t.Partitioned {
		return http.StatusOK, &SubscriptionsResponse{
			Code:      success,
			ErrorCode: tperror.Success,
			Data:      c.service.GetSubscriptionsOfPartitionedTopic(t),
			Detail:    fmt.Sprintf("query subscription success"),
		}
	} else {
		return http.StatusInternalServerError, &SubscriptionsResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorQuerySubscriptionsInfo,
			Message:   c.errMsg.Topic[tperror.ErrorQuerySubscriptionsInfo],
			Data:      nil,
			Detail:    fmt.Sprintf("query subscription error: %+v", err),
		}
	}

}

func (c *controller) SendMessages(req *restful.Request) (int, *SendMessagesResponse) {
	sM := &service.SendMessages{}
	if err := req.ReadEntity(sM); err != nil {
		return http.StatusInternalServerError, &SendMessagesResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorReadEntity,
			Message:   c.errMsg.Topic[tperror.ErrorReadEntity],
			Data:      nil,
			Detail:    fmt.Sprintf("cannot read entity: %+v", err),
		}
	}
	topicId := sM.ID
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &SendMessagesResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Data:      nil,
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	tp, err := c.service.GetTopic(topicId, util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &SendMessagesResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorGetTopicInfo,
			Message:   c.errMsg.Topic[tperror.ErrorGetTopicInfo],
			Detail:    fmt.Sprintf("get database error: %+v", err),
		}
	}
	messageId, err := c.service.SendMessages(tp.URL, sM.MessageBody, sM.Key)
	if err != nil {
		return http.StatusInternalServerError, &SendMessagesResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorSendMessagesError,
			Message:   c.errMsg.Topic[tperror.ErrorSendMessagesError],
			Detail:    fmt.Sprintf("send message error: %+v", err),
		}
	}
	return http.StatusOK, &SendMessagesResponse{
		Code: success,
		Data: messageId,
	}
}

func (c *controller) ResetPosition(req *restful.Request) (int, *ResetPositionResponse) {
	RP := &service.ResetPosition{}
	if err := req.ReadEntity(RP); err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorReadEntity,
			Message:   c.errMsg.Topic[tperror.ErrorReadEntity],
			Detail:    fmt.Sprintf("cannot read entity: %+v", err),
		}
	}
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	tp, err := c.service.GetTopic(RP.ID, util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorGetTopicInfo,
			Message:   c.errMsg.Topic[tperror.ErrorGetTopicInfo],
			Detail:    fmt.Sprintf("get database error: %+v", err),
		}
	}
	if RP.Timestamp > 0 {
		err = c.service.ResetPositionByTime(RP, tp)
	} else {
		err = c.service.ResetPositionById(RP, tp)
	}
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorResetPosition,
			Message:   c.errMsg.Topic[tperror.ErrorResetPosition],
			Detail:    fmt.Sprintf("reset position error: %+v", err),
		}
	} else {
		return http.StatusOK, &ResetPositionResponse{
			Code:    success,
			Message: "success",
		}
	}

}

func (c *controller) SkipAllMessages(req *restful.Request) (int, *ResetPositionResponse) {
	id := req.PathParameter("id")
	subName := req.PathParameter("subName")

	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}

	crd, err := c.service.Get(id, util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorResetPosition,
			Message:   c.errMsg.Topic[tperror.ErrorResetPosition],
			Detail:    fmt.Sprintf("reset cursor error: %+v", err),
		}
	}

	connector := service.NewConnector(c.tpConfig)
	if err = connector.SkipAllMessages(crd, subName); err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorResetPosition,
			Message:   c.errMsg.Topic[tperror.ErrorResetPosition],
			Detail:    fmt.Sprintf("reset cursor error: %+v", err),
		}
	}

	return http.StatusOK, &ResetPositionResponse{
		Code:      success,
		ErrorCode: tperror.Success,
		Message:   "success",
		Detail:    "success",
	}

}

func (c *controller) BatchGrantPermissions(req *restful.Request) (int, *BatchGrantResponse) {
	var topic *service.Topic
	id := req.PathParameter("id")
	BGP := &BatchGrantPermissionRequest{}
	if err := req.ReadEntity(BGP); err != nil {
		return http.StatusInternalServerError, &BatchGrantResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorReadEntity,
			Message:   c.errMsg.Topic[tperror.ErrorReadEntity],
			Detail:    fmt.Sprintf("cannot read entity: %+v", err),
		}
	}
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &BatchGrantResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}

	for _, GrantPermissions := range BGP.ClientAuths {
		if tp, err := c.service.GrantPermissions(id, GrantPermissions.ID, GrantPermissions.Actions, util.WithNamespace(authUser.Namespace)); err != nil {
			return http.StatusInternalServerError, &BatchGrantResponse{
				Code:      fail,
				ErrorCode: tperror.ErrorGrantPermissions,
				Message:   c.errMsg.Topic[tperror.ErrorGrantPermissions],
				Detail:    fmt.Errorf("grant %s permissions error: %+v", GrantPermissions.ID, err).Error(),
			}
		} else {
			topic = tp
		}
	}
	return http.StatusOK, &BatchGrantResponse{
		Code:    success,
		Message: "success",
		Data:    topic,
	}

}

func (c *controller) SkipMessages(req *restful.Request) (int, *ResetPositionResponse) {
	id := req.PathParameter("id")
	subName := req.PathParameter("subName")
	numMessages := req.PathParameter("numMessages")

	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorAuthError,
			Message:   c.errMsg.Topic[tperror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	crd, err := c.service.Get(id, util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorResetPosition,
			Message:   c.errMsg.Topic[tperror.ErrorResetPosition],
			Detail:    fmt.Sprintf("reset cursor error: %+v", err),
		}
	}

	numM, err := strconv.ParseInt(numMessages, 10, 32)
	if err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorResetPosition,
			Message:   c.errMsg.Topic[tperror.ErrorResetPosition],
			Detail:    fmt.Sprintf("reset cursor error: %+v", err),
		}
	}

	connector := service.NewConnector(c.tpConfig)
	if err = connector.SkipMessages(crd, subName, numM); err != nil {
		return http.StatusInternalServerError, &ResetPositionResponse{
			Code:      fail,
			ErrorCode: tperror.ErrorResetPosition,
			Message:   c.errMsg.Topic[tperror.ErrorResetPosition],
			Detail:    fmt.Sprintf("reset cursor error: %+v", err),
		}
	}

	return http.StatusOK, &ResetPositionResponse{
		Code:      success,
		ErrorCode: tperror.Success,
		Message:   "success",
		Detail:    "success",
	}

}
