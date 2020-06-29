/**
Package error提供了Restriction管理的所有错误码（对应ErrorCode字段），
*/
package error
const(
	Success                              				 	= "0"            //成功
	FailedToReadMessageContent								= "007000001" 	//读取消息内容失败
	MessageBodyIsEmpty										= "007000002"	//消息体为空
	FailedToCreateAccessControl								= "007000003"	//创建访问控制失败
	QueryingASingleAccessControlBasedOnIdFails				= "007000004"	//根据id查询单个访问控制失败
	FailedToDeleteAccessControl								= "007000005"	//删除访问控制失败
	QueryAccessControlListFailed							= "007000006"	//查询访问控制列表失败
	QueryAccessControlPagingParameterError					= "007000007"	//查询访问控制分页参数错误
	UpdateAccessControlFailed								= "007000008"	//更新访问控制失败
	BindingOrUnbindingAPIFailed								= "007000009"	//绑定或者解绑api失败
	IncorrectAuthenticationInformation						= "007000010"	//鉴权信息错误
	DuplicateAccessControlName								= "007000011"	//访问控制名字重复
	RequestParameterError									= "007000012"	//请求参数错误
	crdFailedDuringCreation									= "007000013"	//crd创建时失败
)
type RestrictionError struct {
	Err       error
	ErrorCode string
	Message   string
}

func (e *RestrictionError) Error() string { return e.Error() }