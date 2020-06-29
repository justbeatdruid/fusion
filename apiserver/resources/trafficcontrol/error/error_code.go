/**
Package error提供了TrafficControl管理的所有错误码（对应ErrorCode字段），
*/
package error
const(
	Success                              				 	= "0"            //成功
	FailedToReadMessageContent								= "012000001" 	//读取消息内容失败
	MessageBodyIsEmpty										= "012000002"	//消息体为空
	IncorrectAuthenticationInformation                      = "012000003"	//鉴权信息错误
	FailedToCreateTrafficControl							= "012000004"	//创建流量控制失败
	QuerySingleFlowControlFailureBasedOnId					= "012000005"	//根据id查询单个流量控制失败
	FailedToDeleteFlowControl								= "012000006"	//删除流量控制失败
	QueryFlowControlListFailed								= "012000007"	//查询流量控制列表失败
	QueryFlowControlPagingParameterError					= "012000008"	//查询流量控制分页参数错误
	UpdateFlowControlFailed									= "012000009"	//更新流量控制失败
	BindingOrUnbindingAPIFailed								= "012000010"	//绑定或者解绑api失败
	FlowControlWithDuplicateName							= "012000011"	//流量控制名字重复
	RequestParameterError									= "012000012"	//请求参数错误
	crdFailedDuringCreation									= "012000013"	//crd创建时失败
	ErrorsInTheNumberOfAccessesPerUnitTimeLimit				= "012000014"	//每个单位时间限制的访问次数出错
	ThereMustBeAtLeastOneTimeLimit							= "012000015"	//必须存在至少一个时间限制
)
type TrafficControlError struct {
	Err       error
	ErrorCode string
	Message   string
}

func (e *TrafficControlError) Error() string { return e.Error() }