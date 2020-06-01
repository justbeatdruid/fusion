/*

Licensed under the Apache License, Version 2.0 (the "License");
use this fiyou may notUnless required by applicable law or agreed to in writing, software
le except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
limitations unSee the License for the specific language governing permissions and
der the License.
*/

package controllers

import (
	"context"
	"fmt"
	nlptv1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// TopicSynchronizer reconciles a Topic object
type TopicgroupSynchronizer struct {
	client.Client
	Connector *Operator
	clientset.Clientset
}

func (r *TopicgroupSynchronizer) SyncTopicgroup() error {
	klog.Infof("sync topicgroups")
	ctx := context.Background()

	topicGroupList := &nlptv1.TopicgroupList{}

	if err := r.List(ctx, topicGroupList); err != nil {
		return fmt.Errorf("cannot list topics: %+v", err)
	}

	for _, tg := range topicGroupList.Items {
		isExist, err := r.Connector.isNamespacesExist(&tg)
		if err != nil {
			klog.Errorf("pulsar error: %+v", err)
		} else if !isExist {
			tg.Status.Status = nlptv1.Creating
			tg.Status.Message = "sync topic group with apache pulsar"
			if err := r.Update(ctx, &tg); err != nil {
				return fmt.Errorf("cannot sync topic group with apache pulsar: %+v", err)
			}
		}
	}


	return nil
}

func (r *TopicgroupSynchronizer) Start(stop <-chan struct{}) error {
	// wait for cache up
	time.Sleep(time.Second * 10)
	wait.Until(func() {
		if err := r.SyncTopicgroup(); err != nil {
			klog.Errorf( "sync topic stats error: %+v", err)
		}
		// do not use wait.NerverStop
	}, time.Second*60*5, stop)
	return nil
}

// Important
func (*TopicgroupSynchronizer) NeedLeaderElection() bool {
	return true
}
