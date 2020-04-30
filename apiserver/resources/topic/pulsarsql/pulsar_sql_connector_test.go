package pulsarsql

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestConnector_CreateQueryRequest(t *testing.T) {
	c := &Connector{
		Host:       "10.160.32.24",
		Port:       30004,
		PrestoUser: "test-user",
	}

	r, err := c.CreateQueryRequest(`select * from (select row_number() over(order by __publish_time__ desc) row, * from pulsar."public/default"."111" WHERE "__message_id__" = '(191,0,0)') as t where row between 1 and 10`)
	fmt.Printf("result: %+v", r)
	fmt.Println()
	if err != nil {
		t.Error("failed")
	}

	if r.Stats.State == Failed {
		t.Error("failed")
		return
	}

	var index int = 0
	for r.Stats.State != Failed && r.Stats.State != Finished {
		time.Sleep(time.Duration(200) * time.Millisecond)
		url := r.NextUrl
		r, err = c.QueryMessage(r.NextUrl)
		//fmt.Printf("result: %+v", r)

		if err != nil {
			t.Error(fmt.Errorf("failed: %+v", err))
			return
		}

		index++
		fmt.Println(fmt.Sprintf("index:%+v, url: %+v, final result: %+v", index, url, r))

		if len(r.Data) != 0 {
			content, _ := json.Marshal(r.Data)

			fmt.Println(fmt.Sprintf("content: %+v", string(content)))
			return

		}
		if r.Stats.State == Finished {
			//content, _ := json.Marshal(r.Data)
			fmt.Println()
			//fmt.Println(fmt.Sprintf("url: %+v, final result: %+v", url, r))
			//fmt.Println(fmt.Sprintf("content: %+v", string(content)))
			return
		}

	}

}
