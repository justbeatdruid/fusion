package pulsarsql

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
)

func QueryTopicMessages(sql string) ([]service.Messages, error) {
	c := &Connector{
		PrestoUser: "test_user",
		Host:       "10.160.32.24",
		Port:       30004,
	}
	var (
		M     []service.Messages
		ok    bool
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
		} else if state == Finished {
			if response.Data != nil {
				for _, data := range response.Data {
					var m service.Messages
					for k, v := range data {
						switch k {
						case "__message_id__":
							if m.ID, ok = v.(string); !ok {
								return nil, fmt.Errorf("message_id type error")
							}
						case "__publish_time__":
							m.Time, ok = v.(string)
							if !ok {
								return nil, fmt.Errorf("__publish_time__ type error")
							}
						case "__producer_name__":
							if m.ProduceName, ok = v.(string); !ok {
								return nil, fmt.Errorf("__producer_name__ type error")
							}
						case "__partition__":
							if m.Partition, ok = v.(float64); !ok {
								return nil, fmt.Errorf("__partition__ type error")
							}
						case "__key__":
							m.Key = v
						case "__event_time__":
						case "__sequence_id__":
						case "__properties__":
						default:
							m.Messages = append(m.Messages, v)
						}
					}
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