package controllers

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/namespace/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
)

type Operator struct {
	Host string
	Port int
}

const namespaceUrl = "/admin/v2/namespaces/%s/%s"
const protocol = "http"

type requestLogger struct {
	prefix string
}


var logger = &requestLogger{}

func (r *requestLogger) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *requestLogger) Printf(format string, v ...interface{}) {
	klog.V(4).Infof(format, v...)
}

func (r *requestLogger) Println(v ...interface{}) {
	klog.V(4).Infof("%+v", v)
}

func (r *Operator) CreateNamespace (namespace *v1.Namespace) error {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)

	url := fmt.Sprintf(namespaceUrl, namespace.Spec.Tenant, namespace.Spec.Name)
	url = fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, url)

	request.Put(url)
	request.Send("").EndStruct("")
	return nil
}
