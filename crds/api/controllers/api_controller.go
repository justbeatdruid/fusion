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
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	resv1 "github.com/chinamobile/nlpt/crds/restriction/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	trav1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			if api.ObjectMeta.Labels[nlptv1.ApplicationLabel(app.ID)] == "false" && len(app.AclID) != 0 {
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
				api.Status.ApplicationCount = api.Status.ApplicationCount - 1
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
	//遍历绑定api
	if api.Status.Action == nlptv1.Bind {
		api.Status.Status = nlptv1.Running
		for index := 0; index < len(api.Spec.Applications); {
			app := api.Spec.Applications[index]
			if api.ObjectMeta.Labels[nlptv1.ApplicationLabel(app.ID)] == "true" && len(app.AclID) == 0 {
				//consumer添加到白名单中
				aclId, err := r.Operator.AddConsumerToAcl(app.ID, api)
				if err != nil {
					//delete application, ApplicationLabel, Status.Applications
					r.UpdateBindFail(ctx, api, app.ID, index, err.Error())
					//当前删除失败直接返回
					return ctrl.Result{}, nil
				}
				application := &appv1.Application{}
				appNamespacedName := types.NamespacedName{
					Namespace: req.NamespacedName.Namespace,
					Name:      app.ID,
				}
				if err := r.Get(ctx, appNamespacedName, application); err != nil {
					klog.Errorf("cannot get application with name %s: %+v", appNamespacedName, err)
					//delete application, ApplicationLabel, Status.Applications
					r.UpdateBindFail(ctx, api, app.ID, index, err.Error())
					return ctrl.Result{}, nil
				}
				// bind api to application
				api.Spec.Applications[index].AclID = aclId
				api.Status.ApplicationCount = api.Status.ApplicationCount + 1
				if api.Status.Applications == nil {
					api.Status.Applications = make(map[string]nlptv1.ApiApplicationStatus)
				}
				api.Status.Applications[app.ID] = nlptv1.ApiApplicationStatus{
					AppID:            app.ID,
					BindingSucceeded: true,
					Message:          "success",
				}
				application.Spec.APIs = append(application.Spec.APIs, appv1.Api{api.ObjectMeta.Name, api.Spec.Name})
				if err := r.Update(ctx, application); err != nil {
					klog.Errorf("add api to app error: %+v", err)
					//delete application, ApplicationLabel, Status.Applications
					r.UpdateBindFail(ctx, api, app.ID, index, err.Error())
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
		klog.Infof("new api: %s:%s, publish status: %s", api.Spec.Name, api.ObjectMeta.Name, api.Status.PublishStatus)
		if err := r.Operator.CreateRouteByKong(api, su); err != nil {
			api.Status.Status = nlptv1.Error
			api.Status.Message = err.Error()
			r.Update(ctx, api)
			return ctrl.Result{}, nil
		}
		klog.Infof("new api: %s:%s %s", api.Spec.KongApi.PrometheusID, api.Spec.KongApi.AclID, api.Spec.KongApi.JwtID)
		api.Status.Status = nlptv1.Success
		api.Status.PublishStatus = nlptv1.Released
		r.AddApiToServiceUnit(ctx, su, api)
		api.Status.AccessLink = nlptv1.AccessLink(fmt.Sprintf("%s://%s:%d%s",
			strings.ToLower(string(api.Spec.ApiDefineInfo.Protocol)),
			r.Operator.Host, r.Operator.KongPortalPort, api.Spec.KongApi.Paths[0]))
		api.Status.Message = "success"
		r.Update(ctx, api)
		//创建完成后判断是否关联流量控制访问控制，若关联则需要更新相关的状态启动绑定
		if len(api.Spec.Traffic.ID) != 0 {
			r.UpdateTrafficStatus(ctx, req, api, api.Spec.Traffic.ID)
		}
		if len(api.Spec.Traffic.SpecialID) != 0 {
			for _, specialID := range api.Spec.Traffic.SpecialID {
				r.UpdateTrafficStatus(ctx, req, api, specialID)
			}

		}
		if len(api.Spec.Restriction.ID) != 0 {
			r.UpdateRestrictionStatus(ctx, req, api)
		}

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

	//发布API不限制状态  已发布，已下线的API都可以再次发布，未发布的走create流程，其他走update流程
	if api.Status.Action == nlptv1.Publish && api.Status.PublishStatus != nlptv1.UnRelease {
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

	if api.Status.Action == nlptv1.AddPlugins {
		if err := r.Operator.AddResTransformerByKong(api); err != nil {
			klog.Errorf("add rsp transformer by kong err : %+v", err)
			return ctrl.Result{}, nil
		}
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
	for _, value := range su.Spec.APIs {
		if value.ID == api.ObjectMeta.Name {
			klog.Infof("no need update su apis: %+v", *su)
			return nil
		}
	}
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

func (r *ApiReconciler) UpdateBindFail(ctx context.Context, api *nlptv1.Api, appid string, index int, message string) error {
	//delete application, ApplicationLabel, Status.Applications
	api.Spec.Applications = append(api.Spec.Applications[:index], api.Spec.Applications[index+1:]...)
	api.Status.ApplicationCount = api.Status.ApplicationCount - 1
	api.Status.Status = nlptv1.Error
	api.Status.Message = message
	if api.Status.Applications == nil {
		api.Status.Applications = make(map[string]nlptv1.ApiApplicationStatus)
	}
	api.Status.Applications[appid] = nlptv1.ApiApplicationStatus{
		AppID:            appid,
		BindingSucceeded: false,
		Message:          message,
	}
	if err := r.Update(ctx, api); err != nil {
		klog.Errorf("update api error: %+v", err)
		return err
	}
	return nil
}

func (r *ApiReconciler) UpdateTrafficStatus(ctx context.Context, req ctrl.Request, api *nlptv1.Api, id string) {
	tra := &trav1.Trafficcontrol{}
	NamespacedName := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      id,
	}
	if err := r.Get(ctx, NamespacedName, tra); err != nil {
		klog.Errorf("cannot get traffic control with name %s: %+v", NamespacedName, err)
	}
	klog.Infof("get traffic info %+v", *tra)

	tra.Status.Status = trav1.Bind
	tra.ObjectMeta.Labels[api.ObjectMeta.Name] = "true"
	isFisrtBind := true
	for index, v := range tra.Spec.Apis {
		if v.ID == api.ObjectMeta.Name {
			tra.Spec.Apis[index].Result = trav1.BINDING
			isFisrtBind = false
			break
		}
	}
	if isFisrtBind == true {
		tra.Spec.Apis = append(tra.Spec.Apis, trav1.Api{
			ID:       api.ObjectMeta.Name,
			Name:     api.Spec.Name,
			KongID:   api.Spec.KongApi.KongID,
			Result:   trav1.BINDING,
			BindedAt: metav1.Now(),
		})
	}
	klog.Infof("update traffic apis: %+v", tra.Spec.Apis)

	if err := r.Update(ctx, tra); err != nil {
		klog.Errorf("update traffic error: %+v", err)
	}
}

func (r *ApiReconciler) UpdateRestrictionStatus(ctx context.Context, req ctrl.Request, api *nlptv1.Api) {
	res := &resv1.Restriction{}
	NamespacedName := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      api.Spec.Restriction.ID,
	}
	if err := r.Get(ctx, NamespacedName, res); err != nil {
		klog.Errorf("cannot get restriction with name %s: %+v", NamespacedName, err)
	}
	klog.Infof("get restriction info %+v", *res)

	res.Status.Status = resv1.Bind
	res.ObjectMeta.Labels[api.ObjectMeta.Name] = "true"
	isFisrtBind := true
	for index, v := range res.Spec.Apis {
		if v.ID == api.ObjectMeta.Name {
			res.Spec.Apis[index].Result = resv1.BINDING
			isFisrtBind = false
			break
		}
	}
	if isFisrtBind == true {
		res.Spec.Apis = append(res.Spec.Apis, resv1.Api{
			ID:       api.ObjectMeta.Name,
			Name:     api.Spec.Name,
			KongID:   api.Spec.KongApi.KongID,
			Result:   resv1.BINDING,
			BindedAt: util.Now(),
		})
	}
	klog.Infof("update restriction apis: %+v", res.Spec.Apis)
	if err := r.Update(ctx, res); err != nil {
		klog.Errorf("update res error: %+v", err)
	}
}

func (r *ApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Api{}).
		Complete(r)
}
