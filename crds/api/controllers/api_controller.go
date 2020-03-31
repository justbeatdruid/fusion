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
	nlptv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// ApiReconciler reconciles a Api object
type ApiReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=apis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=apis/status,verbs=get;update;patch

const applicationLabel = "application"

func ApplicationLabel(id string) string {
	return strings.Join([]string{applicationLabel, id}, "/")
}

func (r *ApiReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("api", req.NamespacedName)

	api := &nlptv1.Api{}
	if err := r.Get(ctx, req.NamespacedName, api); err != nil {
		klog.Errorf("cannot get api of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	klog.Infof("get new api event: %+v", *api)
	//error 状态不处理  只有Init状态才需要执行
	if api.Status.Status != nlptv1.Init {
		klog.Infof("status is not init no need exec event: %+v", *api)
		return ctrl.Result{}, nil
	}
	//遍历解绑api
	if api.Status.Action == nlptv1.UnBind {
		api.Status.Status = nlptv1.Running
		for index := 0; index < len(api.Spec.Applications); {
			app := api.Spec.Applications[index]
			if api.ObjectMeta.Labels[nlptv1.ApplicationLabel(app.ID)] == "false" {
				//consumer从acl中删除 TODO 失败后处理
				if err := r.Operator.DeleteConsumerFromAcl(app.AclID, app.ID); err != nil {
					api.Status.Status = nlptv1.Error
					api.Status.Message = err.Error()
					r.Update(ctx, api)
					//当前删除失败直接返回
					return ctrl.Result{}, nil
				}
				//delete application, ApplicationLabel, Status.Applications
				api.Spec.Applications = append(api.Spec.Applications[:index], api.Spec.Applications[index+1:]...)
				delete(api.ObjectMeta.Labels, nlptv1.ApplicationLabel(app.ID))
				delete(api.Status.Applications, app.ID)
				application := &appv1.Application{}
				appNamespacedName := types.NamespacedName{
					Namespace: req.NamespacedName.Namespace,
					Name:      app.ID,
				}
				if err := r.Get(ctx, appNamespacedName, application); err != nil {
					klog.Errorf("cannot get application with name %s: %+v", appNamespacedName, err)
					api.Status.Status = nlptv1.Error
					api.Status.Message = err.Error()
					r.Update(ctx, api)
					return ctrl.Result{}, nil
				}
				if err := r.UpdateApp(ctx, application, api); err != nil {
					klog.Errorf("cannot update application with name %s: %+v", appNamespacedName, err)
					api.Status.Status = nlptv1.Error
					api.Status.Message = err.Error()
					r.Update(ctx, api)
					return ctrl.Result{}, nil
				}
			} else {
				index++
			}
		}

		api.Status.Status = nlptv1.Success
		r.Update(ctx, api)
		return ctrl.Result{}, nil
	}

	//first publish   UnRelease--> publish 未发布-->执行发布api-->已发布
	//TODO 下线时是否要删除route   每次发布都重新创建 若删除则API所有的关联都需要重新创建
	if api.Status.Action == nlptv1.Publish && api.Status.PublishStatus == nlptv1.UnRelease {
		// call kong api create
		api.Status.Status = nlptv1.Running
		klog.Infof("new api: %s:%s, publish status: %s", api.Spec.Name, api.ObjectMeta.Name, api.Status.PublishStatus)
		if err := r.Operator.CreateRouteByKong(api); err != nil {
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
			r.Update(ctx, api)
			return ctrl.Result{}, nil
		}
		klog.Infof("new api: %s:%s %s", api.Spec.KongApi.PrometheusID, api.Spec.KongApi.AclID, api.Spec.KongApi.JwtID)
		api.Status.Status = nlptv1.Success
		api.Status.PublishStatus = nlptv1.Released
		//发布成功更新服务单元api的计数
		su := &suv1.Serviceunit{}
		suNamespacedName := types.NamespacedName{
			Namespace: req.NamespacedName.Namespace,
			Name:      api.Spec.Serviceunit.ID,
		}
		if err := r.Get(ctx, suNamespacedName, su); err != nil {
			klog.Errorf("cannot get service unit with name %s: %+v", suNamespacedName, err)
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
			r.Update(ctx, api)
			return ctrl.Result{}, nil
		}
		klog.Infof("get su info %+v", *su)
		r.AddApiToServiceUnit(ctx, su, api)
		api.Status.AccessLink = nlptv1.AccessLink(fmt.Sprintf("%s://%s:%d%s",
			strings.ToLower(string(api.Spec.ApiDefineInfo.Protocol)),
			r.Operator.Host, r.Operator.KongPortalPort, api.Spec.KongApi.Paths[0]))
		api.Status.Message = "success"
		r.Update(ctx, api)
		return ctrl.Result{}, nil
	}

	//下线 api   已经发状态才可以执行下线   已发布-->执行下线api->已下线
	if api.Status.Action == nlptv1.Offline && api.Status.PublishStatus == nlptv1.Released {
		api.Status.Status = nlptv1.Running
		// call kong api delete
		if err := r.Operator.UpdateRouteInfoFromKong(api, true); err != nil {
			//TODO 异常处理
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
			r.Update(ctx, api)
			return ctrl.Result{}, nil

		}
		//PublishStatus -> offline
		api.Status.Status = nlptv1.Success
		api.Status.PublishStatus = nlptv1.Offlined
		su := &suv1.Serviceunit{}
		suNamespacedName := types.NamespacedName{
			Namespace: req.NamespacedName.Namespace,
			Name:      api.Spec.Serviceunit.ID,
		}
		if err := r.Get(ctx, suNamespacedName, su); err != nil {
			klog.Errorf("cannot get service unit with name %s: %+v", suNamespacedName, err)
			return ctrl.Result{}, nil
		}
		klog.Infof("get su info %+v", *su)
		r.RemoveApiFromServiceUnit(ctx, su, api)
		r.Update(ctx, api)
		return ctrl.Result{}, nil
	}

	//发布后的API需要修改 需要先将API下线 之后更新 然后再执行发布  已下线-->执行更新然后发布-->已发布
	if api.Status.Action == nlptv1.Publish && api.Status.PublishStatus == nlptv1.Offlined {
		api.Status.Status = nlptv1.Running
		// call kong api update
		if err := r.Operator.UpdateRouteByKong(api); err != nil {
			//TODO 异常处理
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
			r.Update(ctx, api)
			return ctrl.Result{}, nil
		}
		//PublishStatus -> Released
		api.Status.Status = nlptv1.Success
		api.Status.PublishStatus = nlptv1.Released
		//发布成功更新服务单元api的计数
		su := &suv1.Serviceunit{}
		suNamespacedName := types.NamespacedName{
			Namespace: req.NamespacedName.Namespace,
			Name:      api.Spec.Serviceunit.ID,
		}
		if err := r.Get(ctx, suNamespacedName, su); err != nil {
			klog.Errorf("cannot get service unit with name %s: %+v", suNamespacedName, err)
			return ctrl.Result{}, nil
		}
		klog.Infof("get su info %+v", *su)
		r.AddApiToServiceUnit(ctx, su, api)
		r.Update(ctx, api)
		return ctrl.Result{}, nil
	}
	//未发布API更新 已下线API更新
	if api.Status.Action == nlptv1.Update && (api.Status.PublishStatus == nlptv1.UnRelease ||
		api.Status.PublishStatus == nlptv1.Offlined) {
		api.Status.Status = nlptv1.Success
		klog.Infof("update api : %s %s, status %s", api.Spec.Name, api.ObjectMeta.Name, api.Status.PublishStatus)
		r.Update(ctx, api)
		return ctrl.Result{}, nil
	}
	//已下线删除 已下线时route已经删除
	if api.Status.Action == nlptv1.Delete && api.Status.PublishStatus == nlptv1.Offlined {
		api.Status.Status = nlptv1.Running
		if err := r.Operator.DeleteRouteByKong(api); err != nil {
			//TODO 异常处理
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
			r.Update(ctx, api)
			return ctrl.Result{}, nil

		}
		api.Status.Status = nlptv1.Success
		klog.Infof("delete api : %s %s, status", api.Spec.Name, api.ObjectMeta.Name, api.Status.PublishStatus)
		r.Delete(ctx, api)
		return ctrl.Result{}, nil
	}
	//未发布状态删除  未发布时还未创建route
	if api.Status.Action == nlptv1.Delete && api.Status.PublishStatus == nlptv1.UnRelease {
		api.Status.Status = nlptv1.Success
		klog.Infof("delete api unreleased: %s %s", api.Spec.Name, api.ObjectMeta.Name)
		r.Delete(ctx, api)
		return ctrl.Result{}, nil
	}

	/*if api.Status.Status == nlptv1.Error {
		// call kong api create
		api.Status.Status = nlptv1.Creating
		klog.Infof("error create new api: %s:%d, cafile: %s", r.Operator.Host, r.Operator.Port, r.Operator.CAFile)
		if err := r.Operator.CreateRouteByKong(api); err != nil {
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
		} else {
			api.Status.Status = nlptv1.Created
			api.Status.AccessLink = nlptv1.AccessLink(fmt.Sprintf("%s://%s:%d%s/%s",
				strings.ToLower(string(api.Spec.Protocol)),
				r.Operator.Host, r.Operator.KongPortalPort, api.Spec.KongApi.Paths[0], api.ObjectMeta.Name))
			api.Status.Message = "success"
		}
		r.Update(ctx, api)
	}*/

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ApiReconciler) AddApiToServiceUnit(ctx context.Context, su *suv1.Serviceunit, api *nlptv1.Api) error {
	su.Spec.APIs = append(su.Spec.APIs, suv1.Api{api.ObjectMeta.Name, api.Name})
	su.Status.APICount = su.Status.APICount + 1
	if err := r.Update(ctx, su); err != nil {
		klog.Errorf("update su error: %+v", err)
		return err
	}
	klog.Infof("update su apis: %+v", *su)
	return nil
}
func (r *ApiReconciler) RemoveApiFromServiceUnit(ctx context.Context, su *suv1.Serviceunit, api *nlptv1.Api) error {
	for index, value := range su.Spec.APIs {
		if value.ID == api.ObjectMeta.Name {
			su.Spec.APIs = append(su.Spec.APIs[:index], su.Spec.APIs[index+1:]...)
		}
	}
	su.Status.APICount = su.Status.APICount - 1
	if err := r.Update(ctx, su); err != nil {
		klog.Errorf("update su error: %+v", err)
		return err
	}
	klog.Infof("update su apis: %+v", *su)
	return nil
}

func (r *ApiReconciler) UpdateApp(ctx context.Context, app *appv1.Application, api *nlptv1.Api) error {
	for index, value := range app.Spec.APIs {
		if value.ID == api.ObjectMeta.Name {
			app.Spec.APIs = append(app.Spec.APIs[:index], app.Spec.APIs[index+1:]...)
		}
	}
	if err := r.Update(ctx, app); err != nil {
		klog.Errorf("update app error: %+v", err)
		return err
	}
	klog.Infof("update app apis: %+v", *app)
	return nil
}

func (r *ApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Api{}).
		Complete(r)
}

func (r *ApiReconciler) SyncApiCountFromKong() error {
	klog.Infof("begin sync api count from kong")
	ctx := context.Background()
	apiList := &nlptv1.ApiList{}
	if err := r.List(ctx, apiList); err != nil {
		return fmt.Errorf("cannot list datasources: %+v", err)
	}
	countMap := make(map[string]int)
	if err := r.Operator.syncApiCountFromKong(countMap); err != nil {
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
