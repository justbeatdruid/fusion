package pulsarsql

import (
	"fmt"
	"testing"
)

func TestConnector_CreateQueryRequest(t *testing.T) {
	c := &Connector{
		Host:       "10.160.32.24",
		Port:       30004,
		PrestoUser: "test-user",
	}

	r, err := c.CreateQueryRequest("show Catalogs")
	fmt.Printf("result: %+v", r)
	if err != nil {
		t.Error("failed")
	}

	if r.Stats.State == Failed {
		t.Error("failed")
		return
	}

	for r.Stats.State != Failed && r.Stats.State != Finished {
		r, err = c.QueryMessage(r.NextUrl)
		fmt.Printf("result: %+v", r)
		if err != nil {
			t.Error(fmt.Errorf("failed: %+v", err))
			return
		}
	}

}
