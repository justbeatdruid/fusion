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

func (r *ApiSynchronizer) SyncApiCountFromKong() error {
	klog.Infof("begin sync api count from kong")
	ctx := context.Background()
	apiList := &nlptv1.ApiList{}
	if err := r.List(ctx, apiList); err != nil {
		return fmt.Errorf("cannot list datasources: %+v", err)
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
