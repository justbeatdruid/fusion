package error

//topicgroup错误码：011000001-011111111
const (
	Success                   = "0"
	ErrorReadEntity           = "011000001"
	ErrorCreateTopicgroup     = "011000002"
	ErrorAuthError            = "011000003"
	ErrorBadRequest           = "011000004" //创建Topicgroup失败，参数错误：必填字段不能为空
	ErrorGetTopicgroupList    = "011000005" //查询Topic分组列表失败
	ErrorPageParamInvalid     = "011000006" //查询列表失败：分页参数错误
	ErrorModifyIDInvalid      = "011000007" //修改Topic分组策略失败，原因：未指定Topic分组ID
	ErrorModify               = "011000008" //修改Topic分组策略失败
	ErrorDelete               = "011000009" //删除Topic分组失败
	ErrorDuplicatedTopicgroup = "011000010" //创建Topic分组失败，原因：Topic分组已存在
	ErrorQueryTopicgroupInfo  = "010000011" //查询Topic分组失败

)

type TopicgroupError struct {
	Err       error
	ErrorCode string
	Message   string
}

func (e *TopicgroupError) Error() string { return e.Error() }
