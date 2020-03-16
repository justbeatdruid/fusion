package audit

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/pkg/auth/cas"
	"github.com/chinamobile/nlpt/pkg/logs"

	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
)

type Auditor struct {
	host string
	port int

	queue chan event
}

const path = "/apis/v1/auditlogs"

type event struct {
	data  AutoGenerated
	retry int
}

// NewEvent create a new audit log with tenantID, userID, eventName(delete 4example), eventResult(success 4example)
// resourceType(application 4example), resourceID, resourceName
func (a *Auditor) NewEvent(tenantID, userID, eventName, eventResult, resourceType, resourceID, resourceName string, body string) {
	tenantName := tenantID
	userName, err := cas.GetUserNameByID(userID)
	if err != nil {
		userName = userID
		klog.Errorf("cannot get username with id %s: %+v", userID, err)
	}
	e := event{
		data: AutoGenerated{
			TenantID:     tenantID,
			TenantName:   tenantName,
			UserID:       userID,
			UserName:     userName,
			RecordTime:   time.Now(),
			EventName:    eventName,
			EventResult:  eventResult,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			ResourceName: resourceName,
			RequestBody:  body,
		},
		retry: 2,
	}
	klog.V(5).Infof("new audit log: %+v", e.data)
	a.queue <- e
}

func (a *Auditor) backendLoop(stop <-chan struct{}) {
	klog.Infof("audit backend loop starts with channel length %d", len(a.queue))
	for {
		select {
		case e := <-a.queue:
			if e.retry <= 0 {
				continue
			}
			if err := a.uploadData(e.data); err != nil {
				klog.Errorf("upload auditlog error: %+v", err)
				a.queue <- event{
					data:  e.data,
					retry: e.retry - 1,
				}

			} else {
				klog.V(5).Infof("an audit log uploaded: %+v", e.data)
			}
		}
	}
}

type AutoGenerated struct {
	TenantID     string    `json:"tenantId"`
	TenantName   string    `json:"tenantName"`
	UserID       string    `json:"userId"`
	UserName     string    `json:"userName"`
	RecordTime   time.Time `json:"recordTime"`
	SourceIP     string    `json:"sourceIp"`
	EventName    string    `json:"eventName"`
	EventResult  string    `json:"eventResult"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	ResourceName string    `json:"resourceName"`
	RequestBody  string    `json:"requestBody"`
	ResponseBody string    `json:"responseBody"`
}

type Resp struct {
	Code string `json:"code"`
	Desc string `json:"desc"`
}

func (c *Auditor) uploadData(a AutoGenerated) error {
	request := gorequest.New().SetLogger(logs.GetGoRequestLogger(6)).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, c.host, c.port, path))
	request = request.Set("Content-Type", "application/json")
	request = request.Retry(3, 5*time.Second)

	resp := &Resp{}
	response, body, errs := request.Send(&a).EndStruct(resp)
	if len(errs) > 0 {
		return fmt.Errorf("request for creating audit log error: %+v", errs)
	}
	klog.V(6).Infof("creation response body: %s", string(body))
	if response.StatusCode/100 != 2 {
		klog.V(5).Infof("create operation failed: %d %s", response.StatusCode, string(body))
		return fmt.Errorf("request for quering data error: receive wrong status code: %s", string(body))
	}
	if resp.Code != "0" && resp.Desc != "success" {
		klog.V(5).Infof("create operation failed: received body: %+v", *resp)
		return fmt.Errorf("create operation failed: received body: %+v", *resp)
	}
	return nil
}

func NewAuditor(host string, port int) *Auditor {
	a := &Auditor{
		host:  host,
		port:  port,
		queue: make(chan event, 100),
	}
	//TODO stop received from apiserver/server
	go a.backendLoop(make(chan struct{}, 0))
	return a
}
