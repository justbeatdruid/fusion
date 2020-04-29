/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	nlptv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"k8s.io/klog"
)

// DatasourceSynchronizer reconciles a Datasource object
type ApiSynchronizer struct {
	client.Client
	Operator *Operator
}

func (r *ApiSynchronizer) Start(stop <-chan struct{}) error {
	// wait for cache up
	time.Sleep(time.Second * 10)
	wait.Until(func() {
		if err := r.SyncApiCountFromKong(); err != nil {
			klog.Errorf("sync api count error: %+v", err)
		}
		// do not use wait.NerverStop
	}, time.Second*60, stop)
	return nil
}
func (r *Operator) AddRoutePrometheusByKong(db *nlptv1.Api, id string) error {
	klog.Infof("begin create prometheus for route %s", id)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/routes/", id, "/plugins"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &PrometheusRequestBody{
		Name: "prometheus", //插件名称
	}
	responseBody := &PrometheusResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Infof("add prometheus err: %+v", errs)
		return fmt.Errorf("request for create prometheus error: %+v", errs)
	}
	klog.Infof("creation prometheus code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("create prometheus failed msg: %s\n", responseBody.Message)
		return fmt.Errorf("request for create prometheus error: receive wrong status code: %s", string(body))
	}
	(*db).Spec.KongApi.PrometheusID = responseBody.ID
	klog.Infof("prometheus plugins id is %s", responseBody.ID)
	return nil
}

func (r *ApiSynchronizer) SyncApiCountFromKong() error {
	klog.Infof("begin sync api count from kong")
	ctx := context.Background()
	apiList := &nlptv1.ApiList{}
	if err := r.List(ctx, apiList); err != nil {
		return fmt.Errorf("cannot list datasources: %+v", err)
	}
	for index, value := range apiList.Items {
		//判断api是否配置监控插件，未配置时先添加监控插件
		if len(value.Spec.KongApi.PrometheusID) == 0 {
			klog.Infof("api prometheus id %s", value.Spec.KongApi.PrometheusID)
			apiInfo := &apiList.Items[index]
			//添加监控插件失败不退出打印日志
			if err := r.Operator.AddRoutePrometheusByKong(apiInfo, value.Spec.KongApi.KongID); err != nil {
				klog.Errorf("request for add route prometheus error: %+v", err)
			}
			klog.Infof("api prometheus id %s", value.Spec.KongApi.PrometheusID)
		}
	}
	countMap := make(map[string]int)
	if err := r.Operator.SyncApiCountFromPrometheus(countMap); err != nil {
		return fmt.Errorf("sync api count from kong failed: %+v", err)
	}
	klog.Infof("sync api count map list : %+v", countMap)
	for _, value := range apiList.Items {
		apiID := value.ObjectMeta.Name
		if _, ok := countMap[apiID]; ok {
			if value.Status.CalledCount != countMap[apiID] {
				value.Status.CalledCount = countMap[apiID]
				r.Update(ctx, &value)
			}
		}
	}
	klog.Infof("sync api list %d", len(apiList.Items))
	return nil
}

// Important
func (*ApiSynchronizer) NeedLeaderElection() bool {
	return true
}
