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
	"k8s.io/klog"
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nlptv1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
)

// TopicgroupReconciler reconciles a Topicgroup object
type TopicgroupReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=topicgroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=topicgroups/status,verbs=get;update;patch

func (r *TopicgroupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("topicgroup", req.NamespacedName)

	namespace := &nlptv1.Topicgroup{}
	if err := r.Get(ctx, req.NamespacedName, namespace); err != nil {
		klog.Errorf("cannot get namespace of ctrl req: %+v", err)

		return ctrl.Result{}, nil
	}

	if namespace.Status.Status == nlptv1.Creating {
		//create tenants if it is not exist on Pulsar
		if err := r.Operator.CreateTenantIfNotExist(namespace.ObjectMeta.Namespace); err != nil {
			namespace.Status.Status = nlptv1.CreateFailed
			namespace.Status.Message = err.Error()
			if err := r.Update(ctx, namespace); err != nil {
				klog.Errorf("unable to update namespace: %+v", namespace)
			}
			return ctrl.Result{}, nil
		}

		if err := r.Operator.CreateNamespace(namespace); err != nil {
			namespace.Status.Status = nlptv1.CreateFailed
			namespace.Status.Message = err.Error()
		} else {
			namespace.Status.Status = nlptv1.Created
			namespace.Status.Message = "success"
			namespace.Spec.Available = true
			p, err := r.Operator.GetNamespacePolicies(namespace)
			if err == nil {
				namespace.Spec.Policies = p
				if namespace.Spec.Policies != nil {
					if namespace.Spec.Policies.DeduplicationEnabled == nil {
						*namespace.Spec.Policies.DeduplicationEnabled = false
					}
				}
			} else {

			}

		}

		klog.V(1).Infof("Final Namespace: %+v", *namespace)
		if err := r.Update(ctx, namespace); err != nil {
			klog.Errorf("unable to update namespace: %+v", namespace)
		}
	}

	if namespace.Status.Status == nlptv1.Updating {
		namespace.Status.Message = "updating topic group policies"
		ns, err := r.Operator.GetNamespacePolicies(namespace)
		if err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("get topic group original policies error: %+v", err)

		}

		dstPolicies := namespace.Spec.Policies
		//设置message_ttl_in_seconds
		if ns.MessageTtlInSeconds != dstPolicies.MessageTtlInSeconds {
			if err := r.Operator.SetMessageTTL(namespace); err != nil {
				namespace.Status.Status = nlptv1.UpdateFailed
				namespace.Status.Message = fmt.Sprintf("set message_ttl_in_seconds: %+v", err)

				//设置message_ttl_in_seconds失败，数据回滚
				namespace.Spec.Policies.MessageTtlInSeconds = ns.MessageTtlInSeconds
			}

		}

		//设置retention_polices
		if ns.RetentionPolicies.RetentionTimeInMinutes != dstPolicies.RetentionPolicies.RetentionTimeInMinutes ||
			ns.RetentionPolicies.RetentionSizeInMB != dstPolicies.RetentionPolicies.RetentionSizeInMB {
			if err := r.Operator.SetRetention(namespace); err != nil {
				namespace.Status.Status = nlptv1.UpdateFailed
				namespace.Status.Message = fmt.Sprintf("set retention: %+v", err)

				//设置retention_polices
				namespace.Spec.Policies.RetentionPolicies.RetentionSizeInMB = ns.RetentionPolicies.RetentionSizeInMB
				namespace.Spec.Policies.RetentionPolicies.RetentionTimeInMinutes = ns.RetentionPolicies.RetentionTimeInMinutes
			}
		}

		if (*ns.BacklogQuota)["destination_storage"].Limit != (*dstPolicies.BacklogQuota)["destination_storage"].Limit || (*ns.BacklogQuota)["destination_storage"].Policy != (*dstPolicies.BacklogQuota)["destination_storage"].Policy {
			if err := r.Operator.SetBacklogQuota(namespace); err != nil {
				namespace.Status.Status = nlptv1.UpdateFailed
				namespace.Status.Message = fmt.Sprintf("set backlog quota: %+v", err)

				backlogQuotaMap := make(map[string]nlptv1.BacklogQuota)
				var backlogQuota nlptv1.BacklogQuota
				//设置backlogQuota失败，数据回滚
				backlogQuota.Policy = (*ns.BacklogQuota)["destination_storage"].Policy
				backlogQuota.Limit = (*ns.BacklogQuota)["destination_storage"].Limit
				backlogQuotaMap["destination_storage"] = backlogQuota

			}
		}

		//更新Persistence
		if err = r.Operator.SetPolicy(PersistenceSuffix, namespace, namespace.Spec.Policies.Persistence, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set persistence: %+v", err)
		}

		//更新CompactionThreshold
		if err = r.Operator.SetPolicy(CompactionThresholdSuffix, namespace, namespace.Spec.Policies.CompactionThreshold, put); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set compaction threshold: %+v", err)
		}

		//更新DeduplicationEnable
		if err = r.Operator.SetPolicy(deduplicationSuffix, namespace, namespace.Spec.Policies.DeduplicationEnabled, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set deduplication enable: %+v", err)
		}

		//更新maxConsumersPerSubscription
		if err = r.Operator.SetPolicy(maxConsumersPerSubscriptionSuffix, namespace, namespace.Spec.Policies.MaxConsumersPerSubscription, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set maxConsumersPerSubscription: %+v", err)
		}

		//更新maxConsumersPerTopic
		if err = r.Operator.SetPolicy(maxConsumersPerTopicSuffix, namespace, namespace.Spec.Policies.MaxConsumersPerTopic, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set maxConsumersPerTopicSuffix: %+v", err)
		}

		//更新maxProducersPerTopic
		if err = r.Operator.SetPolicy(maxProducersPerTopicSuffix, namespace, namespace.Spec.Policies.MaxProducersPerTopic, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set maxProducersPerTopic: %+v", err)
		}

		//更新offloadThreshold
		if err = r.Operator.SetPolicy(offloadThresholdSuffix, namespace, namespace.Spec.Policies.OffloadThreshold, put); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set offloadThreshold: %+v", err)
		}

		//更新IsAllowAutoUpdateSchema
		if err = r.Operator.SetPolicy(isAllowAutoUpdateSchemaSuffix, namespace, namespace.Spec.Policies.IsAllowAutoUpdateSchema, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set IsAllowAutoUpdateSchema: %+v", err)
		}

		//更新SchemaCompatibilityStrategy
		//TODO 报错：415
		if err = r.Operator.SetPolicy(schemaCompatibilityStrategySuffix, namespace, strconv.Quote(*namespace.Spec.Policies.SchemaCompatibilityStrategy), put); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set SchemaCompatibilityStrategy: %+v", err)
		}

		//更新SchemaValidationEnforced
		if err = r.Operator.SetPolicy(schemaValidationEnforcedSuffix, namespace, namespace.Spec.Policies.SchemaValidationEnforced, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set SchemaValidationEnforced: %+v", err)
		}

		if namespace.Spec.Policies.OffloadDeletionLagMs != nil {
			//更新OffloadDeletionLagMs
			if err = r.Operator.SetPolicy(offloadDeletionLagMsSuffix, namespace, namespace.Spec.Policies.OffloadDeletionLagMs, put); err != nil {
				namespace.Status.Status = nlptv1.UpdateFailed
				namespace.Status.Message = fmt.Sprintf("set OffloadDeletionLagMs: %+v", err)
			}
		}

		//更新SchemaCompatibilityStrategy
		if err = r.Operator.SetPolicy(encryptionRequiredSuffix, namespace, namespace.Spec.Policies.EncryptionRequired, post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set EncryptionRequired: %+v", err)
		}

		//更新SubscriptionAuthMode,需要在value上加双引号
		if err = r.Operator.SetPolicy(subscriptionAuthModeSuffix, namespace, strconv.Quote(*namespace.Spec.Policies.SubscriptionAuthMode), post); err != nil {
			namespace.Status.Status = nlptv1.UpdateFailed
			namespace.Status.Message = fmt.Sprintf("set SubscriptionAuthMode: %+v", err)
		}

		if namespace.Status.Status == nlptv1.Updating {
			namespace.Status.Status = nlptv1.Updated
			namespace.Status.Message = "modify topic group polices successfully"
		}

		if err = r.Update(ctx, namespace); err != nil {
			klog.Errorf("update database error: %+v", err)
		}

	}

	if namespace.Status.Status == nlptv1.Deleting {
		if err := r.Operator.DeleteNamespace(namespace); err != nil {
			namespace.Status.Status = nlptv1.DeleteFailed
			namespace.Status.Message = err.Error()
			r.Update(ctx, namespace)
		} else {
			r.Delete(ctx, namespace)
		}
		klog.Infof("Final Namespace: %+v", *namespace)

	}

	return ctrl.Result{}, nil
}

func (r *TopicgroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Topicgroup{}).
		Complete(r)
}
