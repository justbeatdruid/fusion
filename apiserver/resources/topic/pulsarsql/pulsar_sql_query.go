package pulsarsql

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
)

func QueryTopicMessages(c Connector, sql string) ([]service.Messages, error) {
	var (
		M     []service.Messages
		state string
	)
	response, err := c.CreateQueryRequest(sql)
	if err != nil {
		return nil, fmt.Errorf(" create query failed: %v ", err)
	}
	//判断状态
	for {
		state = response.Stats.State
		if state == Failed {
			return nil, fmt.Errorf("query failed: %v ", response.Error.Message)
		} else if state == Finished || len(response.Data) > 0 {
			if response.Data != nil {
				for _, data := range response.Data {
					var msg = make(map[string]interface{})
					var m service.Messages
					var size int
					for k, v := range data {
						switch k {
						case "__message_id__":
							m.ID = v
						case "__publish_time__":
							m.Time = v
						case "__producer_name__":
							m.ProducerName = v
						case "__partition__":
							m.Partition = v
						case "__key__":
							m.Key = v
						case "__row__":
						case "__value__":
							//如果是字节数组，需要解码
							value, ok := v.(string)
							if ok {
								decoded, err := base64.StdEncoding.DecodeString(value)
								if err != nil {
									//这种情况发送端直接发的string类型,没有经过base64编码
									m.Message = v
									size = size + binary.Size(v)
								} else {
									m.Message = string(decoded)
									size = size + binary.Size(decoded)
								}
							} else {
								m.Message = v
								size = size + binary.Size(v)
							}
						case "__event_time__":
						case "__sequence_id__":
						case "__properties__":
						case "__count__":
							m.Total = v
						default:
							if v == nil {
								continue
							}
							msg[k] = v
							if str, ok := v.(string); ok {
								size = size + binary.Size([]byte(str))
							} else {
								size = size + binary.Size(v)
							}
						}
					}
					if m.Message == nil {
						m.Message = msg
					}
					m.Size = size
					M = append(M, m)
				}
				return M, nil
			}
			return nil, nil
		} else {
			response, err = c.QueryMessage(response.NextUrl)
			if err != nil {
				return nil, fmt.Errorf("get query failed: %v", err)
			}
		}
	}
}
