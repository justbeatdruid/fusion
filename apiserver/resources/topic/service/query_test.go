package service

import "testing"

func TestConnector_QueryMessage(t *testing.T) {
	QueryTopicMessages(`select * from pulsar."public/default"."111"`)
}
