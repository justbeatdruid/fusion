package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	v1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/parnurzeal/gorequest"
	"net/http"
)

const (
	skipAllUrl    = "/admin/v2/%s/%s/%s/%s/subscriptions/%s/skip_all"
	skillNMessage = "/admin/v2/%s/%s/%s/%s/subscription/%s/skip/%d"
)

//Connector 定义连接Pulsar所需要的参数
type Connector struct {
	Host           string
	Port           int
	AuthEnable     bool
	SuperUserToken string
}

func NewConnector(tpConfig *config.TopicConfig) *Connector {
	return &Connector{
		Host:           tpConfig.Host,
		Port:           tpConfig.Port,
		AuthEnable:     tpConfig.AuthEnable,
		SuperUserToken: tpConfig.AdminToken,
	}
}
func (r *Connector) getHttpRequest() *gorequest.SuperAgent {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	return r.addTokenToHeader(request)

}

func (r *Connector) addTokenToHeader(request *gorequest.SuperAgent) *gorequest.SuperAgent {
	if r.AuthEnable {
		request.Header.Set("Authorization", "Bearer "+r.SuperUserToken)
	}
	return request
}

func (r *Connector) SkipAllMessages(tp *v1.Topic, subscriptionName string) error {

	var domain = "persistent"
	if !tp.Spec.Persistent {
		domain = "non-persistent"
	}

	url := fmt.Sprintf(skipAllUrl, domain, tp.Namespace, tp.Spec.TopicGroup, tp.Spec.Name, subscriptionName)
	url = fmt.Sprintf("%s://%s:%d%s", "http", r.Host, r.Port, url)

	response, body, errs := r.getHttpRequest().Post(url).End()
	if errs != nil {
		return fmt.Errorf("request(%s) failed, error: %+v", url, errs)
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request(%s) failed, response: %+v, body: %+v, error: %+v", url, response, body, errs)
	}

	return nil
}

func (r *Connector) SkipMessages(tp *v1.Topic, subscriptionName string, numMessage int64) error {

	var domain = "persistent"
	if !tp.Spec.Persistent {
		domain = "non-persistent"
	}

	url := fmt.Sprintf(skillNMessage, domain, tp.Namespace, tp.Spec.TopicGroup, tp.Spec.Name, subscriptionName, numMessage)
	url = fmt.Sprintf("%s://%s:%d%s", "http", r.Host, r.Port, url)

	response, body, errs := r.getHttpRequest().Post(url).End()
	if errs != nil {
		return fmt.Errorf("request(%s) failed, error: %+v", url, errs)
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request(%s) failed, response: %+v, body: %+v, error: %+v", url, response, body, errs)
	}

	return nil
}
