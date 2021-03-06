package pulsarsql

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	"strconv"
	"strings"
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
						case "row":
						case "__value__":
							//如果是字节数组，需要解码
							value, ok := v.(string)
							if ok {
								decoded, err := base64.StdEncoding.DecodeString(value)
								if err != nil {
									//这种情况发送端直接发的string类型,没有经过base64编码
									m.Message = value
									size = len(value)
								} else {
									m.Message = string(decoded)
									size = len(string(decoded))
								}
							} else {
								m.Message = v
								size = 0
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
						}
					}

					if m.Message == nil {
						msgJson, err := json.Marshal(msg)
						if err != nil {
							return nil, fmt.Errorf("msg to json error: %s", msgJson)
						}
						msgString := string(msgJson)
						size = len(msgString)
						m.Message = msg
					}

					id := m.ID.(string)
					partition := m.Partition.(float64)
					par := strconv.FormatFloat(partition, 'f', -1, 64)
					ids := strings.Split(id, ",")
					ids = append(ids[0:2], par, ids[2])
					id = strings.Join(ids, ",")
					id = strings.Trim(id, "(")
					id = strings.Trim(id, ")")
					id = strings.ReplaceAll(id, ",", ":")
					m.ID = id
					m.Size = size
					b, _ := json.Marshal(m.Message)
					m.Message = string(b)
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
