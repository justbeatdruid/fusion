/**
Package error提供了Application管理的所有错误码（对应ErrorCode字段），
*/
package error

const(
	Success                              				 	= "0"            //成功
	FailedToReadMessageContent								= "002000001" 	//读取消息内容失败
	MessageBodyIsEmpty										= "002000002"	//消息体为空
	IncorrectAuthenticationInformation                      = "002000003"	//鉴权信息错误
	ApplicationCreationFailed                               = "002000004"	//创建应用失败
	QueryingASingleApplicationBasedOnIdFails                = "002000005"	//根据id查询单个应用失败
	FailedToDeleteApplication                               = "002000006"	//删除应用失败
	UpdateApplicationFailed                                 = "002000007"	//更新应用失败
	QueryApplicationListFailed                              = "002000008"	//查询应用列表失败
	QueryApplicationPagingParameterError                    = "002000009"	//查询应用分页参数错误
	TheIdInTheMessageBodyIsEmpty                            = "002000011"	//消息体中的id为空
	TheRoleInTheMessageBodyIsEmpty                          = "002000012"	//消息体中的角色为空
	FailedToAddUser                                         = "002000013"	//添加用户失败
	TheIdInTheUrlParameterIsEmpty							= "002000014"	//url参数中的id为空
	UserIdInUrlParameterIsEmpty								= "002000015"	//url参数中的用户id为空
	FailedToRemoveUser										= "002000016"	//移除用户失败
	FailedToChangeUser										= "002000017"	//更改用户失败
	FailedToChangeOwner										= "002000018"	//更改拥有者失败
	RequestParameterError									= "002000019"	//请求参数错误
	crdFailedDuringCreation									= "002000020"	//crd创建时失败
	DuplicateApplicationName								= "002000021"	//应用名字重复
)

type ApplicationError struct {
	Err       error
	ErrorCode string
	Message   string
}

func (e *ApplicationError) Error() string { return e.Error() }