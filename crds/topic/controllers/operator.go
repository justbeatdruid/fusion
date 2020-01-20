package controllers

import (
	"errors"
	"fmt"
	nlptv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
)

const persistentTopicUrl = "/admin/v2/persistent/%s/%s/%s"
const protocol = "http"
type requestLogger struct {
	prefix string
}

var logger= &requestLogger{}

func (r *requestLogger) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *requestLogger) Printf(format string, v ...interface{}) {
	klog.V(4).Infof(format, v...)
}

func (r *requestLogger) Println(v ...interface{}) {
	klog.V(4).Infof("%+v", v)
}

//Operator 定义连接Pulsar所需要的参数
type Operator struct{
	Host string
	Port int
}

//CreateTopic 调用Pulsar的Restful Admin API，创建Topic
func (r *Operator) CreateTopic (topic *nlptv1.Topic) (err error){
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)

	klog.Infof("Param: tenant:%s, namespace:%s, topicName:%s", topic.Spec.Tenant, topic.Spec.Namespace, topic.Spec.Name)
	topicUrl := fmt.Sprintf(persistentTopicUrl, topic.Spec.Tenant, topic.Spec.TopicNamespace, topic.Spec.Name)

	topicUrl = fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, topicUrl )
	request = request.Put(topicUrl)
	response, body, errs := request.Send("").EndStruct("")

	fmt.Println("URL:", topicUrl )
	fmt.Print(" Response: ",body, response, errs)

	if response.StatusCode == 204 {
		return nil
	}else {
		errMsg := fmt.Sprintf("Create topic error, url: %s, Error code: %d, Error Message: %s", topicUrl, response.StatusCode, body)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
	return err
}