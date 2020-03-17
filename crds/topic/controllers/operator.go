package controllers

import (
	"errors"
	"fmt"
	nlptv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"time"
)

const persistentTopicUrl = "/admin/v2/persistent/%s/%s/%s"
const nonPersistentTopicUrl = "/admin/v2/non-persistent/%s/%s/%s"
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

//Operator 定义连接Pulsar所需要的参数
type Operator struct {
	Host string
	Port int
	AuthEnable bool
}

//CreateTopic 调用Pulsar的Restful Admin API，创建Topic
func (r *Operator) CreateTopic(topic *nlptv1.Topic) (err error) {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)

	klog.Infof("Param: tenant:%s, namespace:%s, topicName:%s", topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)

	url := persistentTopicUrl
	if topic.Spec.IsNonPersistent {
		url = nonPersistentTopicUrl
	}

	if topic.Spec.Partition > 1 {
		url += "/partitions"
	}
	topicUrl := fmt.Sprintf(url, topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)

	topicUrl = fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, topicUrl)
	request = request.Put(topicUrl)
	response, body, errs := request.Send("").EndStruct("")

	fmt.Println("URL:", topicUrl)
	fmt.Println(" Response: ", body, response, errs)

	if response.StatusCode == 204 {
		return nil
	} else {
		errMsg := fmt.Sprintf("Create topic error, url: %s, Error code: %d, Error Message: %s", topicUrl, response.StatusCode, body)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

//DeleteTopic 调用Pulsar的Restful Admin API，删除Topic
func (r *Operator) DeleteTopic(topic *nlptv1.Topic) (err error) {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	url := persistentTopicUrl
	if topic.Spec.IsNonPersistent {
		url = nonPersistentTopicUrl
	}

	if topic.Spec.Partition > 1 {
		url += "/partitions"
	}
	topicUrl := fmt.Sprintf(url, topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)

	topicUrl = fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, topicUrl)
	request = request.Delete(topicUrl)
	response, body, errs := request.Send("").EndStruct("")
	fmt.Println("URL:", topicUrl)
	fmt.Print(" Response: ", body, response, errs)
	if response.StatusCode == 204 {
		return nil
	} else {
		errMsg := fmt.Sprintf("delete topic error, url: %s, Error code: %d, Error Message: %s", topicUrl, response.StatusCode, body)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

//删除授权
func (r *Operator) DeletePer(topic *nlptv1.Topic,P *nlptv1.Permission) (err error) {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	url := persistentTopicUrl
	if topic.Spec.IsNonPersistent {
		url = nonPersistentTopicUrl
	}

	if topic.Spec.Partition > 1 {
		url += "/partitions"
	}
	topicUrl := fmt.Sprintf(url, topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)
	topicUrl = fmt.Sprintf("%s://%s:%d%s%s%s", protocol, r.Host, r.Port, topicUrl,"permissions",P.AuthUserName)
	request = request.Delete(topicUrl).Retry(3, 5*time.Second)
	response, body, errs := request.Send("").EndStruct("")
	fmt.Println("URL:", topicUrl)
	fmt.Print(" Response: ", body, response, errs)
	if response.StatusCode == 204 {
		return nil
	} else {
		errMsg := fmt.Sprintf("delete topic error, url: %s, Error code: %d, Error Message: %s", topicUrl, response.StatusCode, body)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}
