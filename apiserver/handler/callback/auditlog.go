package callback

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/chinamobile/nlpt/pkg/audit"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/go-restful"

	"k8s.io/klog"
)

const (
	GET    = "GET"
	DELETE = "DELETE"
	PATCH  = "PATCH"
	POST   = "POST"
	PUT    = "PUT"
)

type selector struct {
	route  string
	method string
}

var accepted = []selector{
	{"/api/v1/apis", POST},
	{"/api/v1/applications", POST},
	{"/api/v1/applies", POST},
	{"/api/v1/datasources", POST},
	{"/api/v1/serviceunits", POST},
	{"/api/v1/apis/{id}", DELETE},
	{"/api/v1/applications/{id}", DELETE},
	{"/api/v1/applies/{id}", DELETE},
	{"/api/v1/datasources/{id}", DELETE},
	{"/api/v1/serviceunits/{id}", DELETE},
}

// return event, resource and if this request should be uploaded as event
func filter(req *restful.Request) (string, string, bool) {
	for _, a := range accepted {
		if req.Request.Method == a.method && req.SelectedRoutePath() == a.route {
			return getEventName(a.method), getResourceType(a.route), true
		}
	}
	return "", "", false
}

func getResourceType(path string) string {
	ss := strings.Split(path, "/")
	if len(ss) > 3 {
		switch strings.ToLower(ss[3]) {
		case "apis":
			return "API"
		case "applications":
			return "应用"
		case "applies":
			return "申请工单"
		case "datasources":
			return "数据源"
		case "serviceunits":
			return "服务单元"
		}
	}
	return "未知"
}

func getEventName(method string) string {
	switch strings.ToUpper(method) {
	case "POST":
		return "创建"
	case "PATCH", "PUT":
		return "更新"
	case "DELETE":
		return "删除"
	case "GET":
		return "查询"
	default:
		return "未知"
	}
}

type Resource struct {
	Data *Data `json:"data,omitempty"`
}

type Data struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getResourceFromEntity(entity interface{}) (*Resource, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity is null")
	}
	b, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %+v", err)
	}
	r := &Resource{}
	err = json.Unmarshal(b, r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %+v", err)
	}
	if r.Data == nil {
		return nil, fmt.Errorf("cannot unmarshal data in %s", string(b))
	}
	return r, nil
}

func NewAuditCaller(c *restful.Container, a *audit.Auditor) func(*restful.Request, *restful.Response, *restful.CallbackChain) {
	return func(req *restful.Request, resp *restful.Response, chain *restful.CallbackChain) {
		eventName, resourceType, ok := filter(req)
		if !ok {
			chain.ProcessCallback(req, resp)
			return
		}

		user, err := auth.GetAuthUser(req)
		if err != nil {
			klog.Errorf(err.Error())
			chain.ProcessCallback(req, resp)
			return
		}
		tenantID, userID := user.Namespace, user.Name
		var eventResult string
		switch resp.StatusCode() / 100 {
		case 2:
			eventResult = "Success"
		case 3:
			eventResult = "Retring"
		case 4:
			eventResult = "Error"
		case 5:
			eventResult = "Failed"
		default:
			eventResult = "Unknown"
		}
		var resourceID, resourceName string
		resourceID = req.PathParameter("id")
		var mutex sync.Mutex
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			entities := resp.GetEntities()
			if !(len(entities) > 0) {
				klog.Errorf("cannot get one entity from resp")
				return
			}
			if entities[0] == nil {
				return
			}
			r, err := getResourceFromEntity(entities[0])
			if err != nil {
				klog.Errorf("cannot get resource from response entity: %+v", err)
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			if len(resourceID) == 0 && r.Data != nil {
				resourceID = r.Data.ID
			}
			if len(resourceName) == 0 && r.Data != nil {
				resourceName = r.Data.Name
			}
		}()
		go func() {
			defer wg.Done()
			entity := req.GetEntity()
			if entity == nil {
				return
			}
			r, err := getResourceFromEntity(entity)
			if err != nil {
				klog.Errorf("cannot get resource from request entity: %+v", err)
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			if len(resourceID) == 0 && r.Data != nil {
				resourceID = r.Data.ID
			}
			if len(resourceName) == 0 && r.Data != nil {
				resourceName = r.Data.Name
			}
		}()
		wg.Wait()

		a.NewEvent(tenantID, userID, eventName, eventResult, resourceType, resourceID, resourceName)
		chain.ProcessCallback(req, resp)
	}
}
