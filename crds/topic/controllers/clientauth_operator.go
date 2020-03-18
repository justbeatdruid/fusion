package controllers

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
	"time"
)

const (
	Nofity_URL = "api/v1/clientauths/%s/topics/%s"
)

type ClientAuthOperator struct {
	Host     string
	Port     int
	Protocol string
}

func (c *ClientAuthOperator) AddAuthorizedTopic(authId string, topicId string) error {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)

	url := fmt.Sprintf(Nofity_URL, authId, topicId)
	url = fmt.Sprintf("%s://%s:%d/%s", c.Protocol, c.Host, c.Port, url)

	response, body, errs := request.Put(url).Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError).Set("content-type", "application/json").End()
	if response.StatusCode == http.StatusOK {
		return nil
	}

	klog.Errorf("add authorized topic error, url:%+v, response :%+v, body: %+v, err:%+v", url, response, body, errs)
	return fmt.Errorf("add authorized topic error, err:%+v", errs)

}

func (c *ClientAuthOperator) DeleteAuthorizedTopic(authId string, topicId string) error {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)

	url := fmt.Sprintf(Nofity_URL, authId, topicId)
	url = fmt.Sprintf("%s://%s:%d/%s", c.Protocol, c.Host, c.Port, url)

	response, body, errs := request.Delete(url).Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError).Set("content-type", "application/json").End()
	if response.StatusCode == http.StatusOK {
		return nil
	}

	klog.Errorf("remove authorized topic error, url:%+v, response :%+v, body: %+v, err:%+v", url, response, body, errs)
	return fmt.Errorf("remove authorized topic error, err:%+v", errs)

}
