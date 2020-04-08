package pulsar

const (
	DefaultNumberOfNamespaceBundles = 4 //Pulsar中默认的Namespace Bundle数量
	DefaultMessageTTlInSeconds      = 0 //Pulsar中未确认消息的最长保留时间，超过此时间消息会被自动确认
	DefaultBacklogPolicy            = "producer_request_hold"
	DefaultRetentionTimeInMinutes   = 24 * 3 * 60 //默认保留三天的数据
	DefaultRetentionSizeInMB        = 0
	MinRetentionTimeInMinutes       = -1 //-1代表unlimited
	MinRetentionSizeInMB            = -1 //-1代表unlimited
)
