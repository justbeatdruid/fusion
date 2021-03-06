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
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	nlptv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"

	"k8s.io/klog"
)

//var defaultNamespace = "default"

// TopicSynchronizer reconciles a Topic object
type TopicSynchronizer struct {
	client.Client
	Connector *Connector
}

func (r *TopicSynchronizer) SyncTopicStats() error {
	klog.Infof("sync topics stats")
	ctx := context.Background()

	topicList := &nlptv1.TopicList{}
	if err := r.List(ctx, topicList); err != nil {
		return fmt.Errorf("cannot list topics: %+v", err)
	}
	for _, tp := range topicList.Items {
		if tp.Status.Status == nlptv1.Creating {
			continue
		}

		go r.syncStats(tp)

	}

	return nil
}

func (r *TopicSynchronizer) syncStats(tp nlptv1.Topic) {
	ctx := context.Background()
	stats, err := r.Connector.GetStats(tp)
	if err != nil {
		return
	}
	tp.Spec.Stats = *stats
	if err := r.Update(ctx, &tp); err != nil {
		klog.Errorf("update topic stats error")
	} else {
		klog.Info("finished sync stats of topic:", tp.Spec.Name)
	}
}
func (r *TopicSynchronizer) Start(stop <-chan struct{}) error {
	// wait for cache up
	time.Sleep(time.Second * 10)
	wait.Until(func() {
		if err := r.SyncTopicStats(); err != nil {
			klog.Errorf("sync topic stats error: %+v", err)
		}
		// do not use wait.NerverStop
	}, time.Second*60, stop)
	return nil
}

// Important
func (*TopicSynchronizer) NeedLeaderElection() bool {
	return true
}
