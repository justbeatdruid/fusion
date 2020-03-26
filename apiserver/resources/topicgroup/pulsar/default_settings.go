package pulsar

const (
	DefaultNumberOfNamespaceBundles = 4 //Pulsar中默认的Namespace Bundle数量
	DefaultMessageTTlInSeconds      = 0 //Pulsar中未确认消息的最长保留时间，超过此时间消息会被丢弃，默认不保留
	DefaultBacklogPolicy            = "producer_request_hold"
	DefaultRetentionTimeInMinutes   = 0
	DefaultRetentionSizeInMB        = 0
)
