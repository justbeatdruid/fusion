package callback

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/pkg/go-restful"

	"k8s.io/klog"
)

type Wrapped struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func getWrappedFromEntity(entity interface{}) (Wrapped, error) {
	if entity == nil {
		return Wrapped{}, fmt.Errorf("entity is null")
	}
	b, err := json.Marshal(entity)
	if err != nil {
		return Wrapped{}, fmt.Errorf("marshal error: %+v", err)
	}
	r := Wrapped{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return Wrapped{}, fmt.Errorf("unmarshal error: %+v", err)
	}
	return r, nil
}

func NewLogger() func(*restful.Request, *restful.Response) {
	return func(req *restful.Request, resp *restful.Response) {
		entities := resp.GetEntities()
		if !(len(entities) > 0) {
			return
		}
		if entities[0] == nil {
			return
		}
		w, err := getWrappedFromEntity(entities[0])
		if err != nil {
			return
		}
		if w.Code != 0 {
			if req.Request.URL == nil {
				return
			}
			klog.Errorf("fusion Rest API error: path %s, method %s, detail: %s", req.Request.URL.String(), req.Request.Method, w.Detail)
		}
	}
}
