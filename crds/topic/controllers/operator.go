package controllers

import (
	"errors"
	"fmt"
	nlptv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
	"time"
)

const (
	persistentTopicUrl         = "/admin/v2/persistent/%s/%s/%s"
	nonPersistentTopicUrl      = "/admin/v2/non-persistent/%s/%s/%s"
	protocol                   = "http"
	persistentPermissionUrl    = "/admin/v2/persistent/%s/%s/%s/permissions/%s"
	nonPersistentPermissionUrl = "/admin/v2/non-persistent/%s/%s/%s/permissions/%s"
)

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
	Host           string
	Port           int
	AuthEnable     bool
	SuperUserToken string
}

//CreateTopic 调用Pulsar的Restful Admin API，创建Topic
func (r *Operator) CreateTopic(topic *nlptv1.Topic) (err error) {
	if topic.Spec.Partition > 1 {
		return r.CreatePartitionedTopic(topic)
	}

	request := r.GetHttpRequest()
	klog.Infof("Param: tenant:%s, namespace:%s, topicName:%s", topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)
	topicUrl := r.getUrl(topic)
	response, _, errs := request.Put(topicUrl).Send("").EndStruct("")
	if response.StatusCode == 204 {
		return nil
	} else {
		errMsg := fmt.Sprintf("Create topic error, url: %s, Error code: %d, Error Message: %+v", topicUrl, response.StatusCode, errs)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

func (r *Operator) CreatePartitionedTopic(topic *nlptv1.Topic) (err error) {
	request := r.GetHttpRequest()
	klog.Infof("CreatePartitionedTopic Param: tenant:%s, namespace:%s, topicName:%s", topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)
	topicUrl := r.getUrl(topic)

	response, _, errs := request.Put(topicUrl).Send(topic.Spec.Partition - 1).EndStruct("")
	if response.StatusCode == 204 {
		return nil
	} else {
		errMsg := fmt.Sprintf("Create topic error, url: %s, Error code: %d, Error Message: %+v", topicUrl, response.StatusCode, errs)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

//DeleteTopic 调用Pulsar的Restful Admin API，删除Topic
func (r *Operator) DeleteTopic(topic *nlptv1.Topic) (err error) {
	request := r.GetHttpRequest()
	topicUrl := r.getUrl(topic)
	response, body, errs := request.Delete(topicUrl).Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError).End()
	fmt.Println("URL:", topicUrl)
	fmt.Print(" Response: ", body, response, errs)
	if response.StatusCode == 204 {
		return nil
	} else if body == "Topic not found" || body == "Partitioned topic does not exist" {
		return nil
	} else {
		errMsg := fmt.Sprintf("delete topic error, url: %s, Error code: %d, Error Message: %s", topicUrl, response.StatusCode, body)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

func (r *Operator) GrantPermission(topic *nlptv1.Topic, permission *nlptv1.Permission) (err error) {
	request := r.GetHttpRequest()
	var url string
	if topic.Spec.IsNonPersistent {
		url = nonPersistentPermissionUrl
	} else {
		url = persistentPermissionUrl
	}

	url = fmt.Sprintf(url, topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name, permission.AuthUserName)
	url = fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, url)
	response, body, errs := request.Post(url).Send(permission.Actions).End()

	klog.Infof("grant permission result, url: %+v, response: %+v, body: %+v, err:%+v", url, response, body, errs)
	if response.StatusCode == 204 {
		return nil
	}

	return fmt.Errorf("grant permission error: %+v", errs)

}

func (r *Operator) getUrl(topic *nlptv1.Topic) string {
	url := persistentTopicUrl
	if topic.Spec.IsNonPersistent {
		url = nonPersistentTopicUrl
	}

	if topic.Spec.Partition > 1 {
		url += "/partitions"
	}
	topicUrl := fmt.Sprintf(url, topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)

	return fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, topicUrl)
}

//删除授权
func (r *Operator) DeletePer(topic *nlptv1.Topic, P *nlptv1.Permission) (err error) {
	request := r.GetHttpRequest()
	url := persistentTopicUrl
	if topic.Spec.IsNonPersistent {
		url = nonPersistentTopicUrl
	}

	topicUrl := fmt.Sprintf(url, topic.Spec.Tenant, topic.Spec.TopicGroup, topic.Spec.Name)
	topicUrl = fmt.Sprintf("%s://%s:%d%s%s%s", protocol, r.Host, r.Port, topicUrl, "permissions", P.AuthUserName)
	response, _, errs := request.Delete(topicUrl).Retry(3, 5*time.Second).Send("").EndStruct("")

	if response.StatusCode == 204 {
		return nil
	}
	errMsg := fmt.Sprintf("delete topic error, url: %s, Error code: %d, Error Message: %+v", topicUrl, response.StatusCode, errs)
	klog.Error(errMsg)
	return errors.New(errMsg)
}

func (r *Operator) AddTokenToHeader(request *gorequest.SuperAgent) *gorequest.SuperAgent {
	if r.AuthEnable {
		request.Header.Set("Authorization", "Bearer "+r.SuperUserToken)
	}
	return request
}

func (r *Operator) GetHttpRequest() *gorequest.SuperAgent {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	return r.AddTokenToHeader(request)

}
