/**
Package error提供了Topic管理的所有错误码（对应ErrorCode字段），
*/
package error

//topic错误码：010000001-010111111
const (
	Success                        = "0"         //成功
	Error_Read_Entity              = "010000001" //读取消息体失败
	Error_Create_Topic             = "010000002" //创建Topic失败
	Error_Auth_Error               = "010000003" //鉴权信息获取失败
	Error_Bad_Request              = "010000004" //创建Topic失败，参数错误：必填字段不能为空
	Error_Topic_Exists             = "010000005" //创建Topic失败，Topic已存在
	Error_Get_TopicInfo            = "010000006" //查询Topic详情失败
	Error_Delete_Topic             = "010000007" //删除Topic失败
	Error_Get_TopicList            = "010000008" //查询列表失败
	Error_Page_Param_Invalid       = "010000009" //查询列表失败：分页参数错误
	Error_Parse_Import_File        = "010000010" //导入Topic失败：文件解析错误
	Error_Import_Bad_Request       = "010000011" //导入Topic失败：参数校验错误
	Error_Query_Message            = "010000012" //查询消息失败
	Error_Query_Message_StartTime  = "010000013" //查询消息失败，原因：起始时间错误
	Error_Query_Message_EndTime    = "010000014" //查询消息失败，原因：结束时间错误
	Error_Query_Message_Page_Param = "010000015" // 查询消息失败，原因：分页参数错误
	Error_Grant_Permissions        = "010000016" //用户授权失败
	Error_Revoke_Permissions       = "010000017" //收回用户权限失败

)

type TopicError struct {
	Err       error
	ErrorCode string
	Message   string
}

func (e *TopicError) Error() string { return e.Error() }
