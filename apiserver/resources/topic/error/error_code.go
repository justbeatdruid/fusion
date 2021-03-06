/**
Package error提供了Topic管理的所有错误码（对应ErrorCode字段），
*/
package error

//topic错误码：010000001-010111111
const (
	Success                               = "0"         //成功
	ErrorReadEntity                       = "010000001" //读取消息体失败
	ErrorCreateTopic                      = "010000002" //创建Topic失败
	ErrorAuthError                        = "010000003" //鉴权信息获取失败
	ErrorBadRequest                       = "010000004" //创建Topic失败，参数错误：必填字段不能为空
	ErrorTopicExists                      = "010000005" //创建Topic失败，Topic已存在
	ErrorGetTopicInfo                     = "010000006" //查询Topic详情失败
	ErrorDeleteTopic                      = "010000007" //删除Topic失败
	ErrorGetTopicList                     = "010000008" //查询列表失败
	ErrorPageParamInvalid                 = "010000009" //查询列表失败：分页参数错误
	ErrorParseImportFile                  = "010000010" //导入Topic失败：文件解析错误
	ErrorImportBadRequest                 = "010000011" //导入Topic失败：参数校验错误
	ErrorQueryMessage                     = "010000012" //查询消息失败
	ErrorQueryMessageStartTime            = "010000013" //查询消息失败，原因：起始时间错误
	ErrorQueryMessageEndTime              = "010000014" //查询消息失败，原因：结束时间错误
	ErrorQueryMessagePageParam            = "010000015" //查询消息失败，原因：分页参数错误
	ErrorGrantPermissions                 = "010000016" //用户授权失败
	ErrorRevokePermissions                = "010000017" //收回用户权限失败
	ErrorImportTopics                     = "010000018" //导入失败
	ErrorEnsureNamespace                  = "010000019" //确认k8s命名空间错误
	ErrorCannotFindTopicgroup             = "010000020" //创建Topic失败，原因：找不到Topic分组
	ErrorCannotFindTopic                  = "010000021" //查询信息失败，原因：找不到Topic
	ErrorTopicIdError                     = "010000022" //查询消息失败：原因：topic字段不能为空
	ErrorTopicGroupIdError                = "010000023" //查询消息失败：原因：topicGroup字段不能为空
	ErrorGetTopicGroupInfo                = "010000024" //查询TopicGroup详情失败
	ErrorQueryParameterError              = "010000025" //查询信息失败，原因：查询参数不正确
	ErrorPartitionTopicPartitionEqualZero = "010000026" //创建Topic失败，原因：多分区Topic的分区数只能为1～20之间的整数
	ErrorAddPartitionsOfTopicError        = "010000027" //增加topic分区失败
	ErrorBindTopicError                   = "010000028" //应用与Topic绑定/解绑定失败
	ErrorSendMessagesError                = "010000029" //发送消息失败
	ErrorQuerySubscriptionsInfo           = "010000030" //查询Topic订阅关系失败
	ErrorUnBindTopicError                 = "010000031" //应用与Topic解绑定失败
	ErrorModifyPermissions                = "010000032" //修改权限失败
	ErrorResetPosition                    = "010000033" //重置消费位置失败
	ErrorRefresh                          = "010000034" //刷新失败
	ErrorImport                           = "010000035" //导入失败，原因：%+v
	ErrorModifyDescription                = "010000036" //修改描述失败，原因：%+v
	ErrorForceDelete                      = "010000037" //强制删除失败，原因：%+v
	ErrorTerminate                        = "010000038" //终止失败，原因：Topic不存在
)

type TopicError struct {
	Err       error
	ErrorCode string
	Message   string
}

func (e *TopicError) Error() string { return e.Error() }
