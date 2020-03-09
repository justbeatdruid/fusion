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
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
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
				//delete application
				api.Spec.Applications = append(api.Spec.Applications[:index], api.Spec.Applications[index+1:]...)
				delete(api.ObjectMeta.Labels, nlptv1.ApplicationLabel(app.ID))
				r.Update(ctx, api)
			} else {
				index++
			}
		}
		api.Status.Status = nlptv1.Success
		return ctrl.Result{}, nil
	}

	//first publish   UnRelease--> publish 未发布-->执行发布api-->已发布
	if api.Status.Action == nlptv1.Publish && api.Status.PublishStatus == nlptv1.UnRelease {
		// call kong api create
		api.Status.Status = nlptv1.Running
		klog.Infof("new api: %s:%d, cafile: %s", r.Operator.Host, r.Operator.Port, r.Operator.CAFile)
		if err := r.Operator.CreateRouteByKong(api); err != nil {
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
			r.Update(ctx, api)
			return ctrl.Result{}, nil
		}
		if api.Spec.AuthType == nlptv1.APPAUTH {
			//在route上创建jwt及acl插件
			if err := r.Operator.AddRouteJwtByKong(api); err != nil {
				api.Status.Status = nlptv1.Error
				api.Status.Message = err.Error()
				r.Update(ctx, api)
				return ctrl.Result{}, nil
			}
			if err := r.Operator.AddRouteAclByKong(api); err != nil {
				api.Status.Status = nlptv1.Error
				api.Status.Message = err.Error()
				r.Update(ctx, api)
				return ctrl.Result{}, nil
			}
		}
		if api.Spec.ApiDefineInfo.Cors == "true" {
			if err := r.Operator.AddRouteCorsByKong(api); err != nil {
				api.Status.Status = nlptv1.Error
				api.Status.Message = err.Error()
				r.Update(ctx, api)
				return ctrl.Result{}, nil
			}
		}
		api.Status.Status = nlptv1.Success
		api.Status.PublishStatus = nlptv1.Released
		api.Status.AccessLink = nlptv1.AccessLink(fmt.Sprintf("%s://%s:%d%s",
			strings.ToLower(string(api.Spec.ApiDefineInfo.Protocol)),
			r.Operator.Host, r.Operator.KongPortalPort, api.Spec.KongApi.Paths[0]))
		api.Status.Message = "success"
		r.Update(ctx, api)
	}

	//下线 api   已经发状态才可以执行下线   已发布-->执行下线api->已下线
	if api.Status.Action == nlptv1.Offline && api.Status.PublishStatus == nlptv1.Released {
		api.Status.Status = nlptv1.Running
		// call kong api delete
		if err := r.Operator.DeleteRouteByKong(api); err != nil {
			//TODO 异常处理
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
		}
		//PublishStatus -> offline
		api.Status.Status = nlptv1.Success
		api.Status.PublishStatus = nlptv1.Offlined
		r.Update(ctx, api)
	}

	//发布后的API需要修改 需要先将API下线 之后更新 然后再执行发布  已下线-->执行更新然后发布-->已发布
	if api.Status.Action == nlptv1.Publish && api.Status.PublishStatus == nlptv1.Offlined {
		api.Status.Status = nlptv1.Running
		// call kong api update
		if err := r.Operator.UpdateRouteByKong(api); err != nil {
			//TODO 异常处理
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
		}
		//PublishStatus -> Released
		api.Status.Status = nlptv1.Success
		api.Status.PublishStatus = nlptv1.Released
		r.Update(ctx, api)
	}

	if api.Status.Action == nlptv1.Delete {
		api.Status.Status = nlptv1.Running
		// call kong api delete
		if err := r.Operator.DeleteRouteByKong(api); err != nil {
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
		}
		api.Status.Status = nlptv1.Success
		r.Delete(ctx, api)
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

func (r *ApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Api{}).
		Complete(r)
}
