/**
Package error提供了Topic管理的所有错误码（对应ErrorCode字段），
*/
package error

//topic错误码：010000001-010111111
const (
	Success                    = "0"         //成功
	ErrorReadEntity            = "010000001" //读取消息体失败
	ErrorCreateTopic           = "010000002" //创建Topic失败
	ErrorAuthError             = "010000003" //鉴权信息获取失败
	ErrorBadRequest            = "010000004" //创建Topic失败，参数错误：必填字段不能为空
	ErrorTopicExists           = "010000005" //创建Topic失败，Topic已存在
	ErrorGetTopicInfo          = "010000006" //查询Topic详情失败
	ErrorDeleteTopic           = "010000007" //删除Topic失败
	ErrorGetTopicList          = "010000008" //查询列表失败
	ErrorPageParamInvalid      = "010000009" //查询列表失败：分页参数错误
	ErrorParseImportFile       = "010000010" //导入Topic失败：文件解析错误
	ErrorImportBadRequest      = "010000011" //导入Topic失败：参数校验错误
	ErrorQueryMessage          = "010000012" //查询消息失败
	ErrorQueryMessageStartTime = "010000013" //查询消息失败，原因：起始时间错误
	ErrorQueryMessageEndTime   = "010000014" //查询消息失败，原因：结束时间错误
	ErrorQueryMessagePageParam = "010000015" // 查询消息失败，原因：分页参数错误
	ErrorGrantPermissions      = "010000016" //用户授权失败
	ErrorRevokePermissions     = "010000017" //收回用户权限失败

)

type TopicError struct {
	Err       error
	ErrorCode string
	Message   string
}

func (e *TopicError) Error() string { return e.Error() }
