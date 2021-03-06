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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	nlptv1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	dwv1 "github.com/chinamobile/nlpt/crds/datasource/datawarehouse/api/v1"
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
	"github.com/chinamobile/nlpt/pkg/names"

	"k8s.io/klog"
)

// DatasourceSynchronizer reconciles a Datasource object
type DatasourceSynchronizer struct {
	client.Client
	DataConnector dw.Connector
}

func GenerateName(db *dwv1.Database) string {
	if db == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", db.Name, db.SubjectName)
}

func GetDataWarehouseKey(db *dwv1.Database) string {
	if db == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", db.Id, db.SubjectId)
}

func (r *DatasourceSynchronizer) SyncDatasources() error {
	klog.Infof("sync datasources")
	ctx := context.Background()

	apiDatasourceList := &nlptv1.DatasourceList{}
	if err := r.List(ctx, apiDatasourceList, &client.ListOptions{
		Namespace: defaultNamespace,
	}); err != nil {
		return fmt.Errorf("cannot list datasources: %+v", err)
	}
	existedDatawarehouses := make(map[string]nlptv1.Datasource)
	for i, ds := range apiDatasourceList.Items {
		klog.V(6).Infof("get datasources: %dth datasource: %+v", i, ds)
		if ds.Spec.Type == nlptv1.DataWarehouseType {
			existedDatawarehouses[GetDataWarehouseKey(ds.Spec.DataWarehouse)] = ds
		}
	}
	datawarehouse, err := r.DataConnector.GetExampleDatawarehouse() //查询新的数据
	if err != nil {
		return fmt.Errorf("get datawarehouse error: %+v", err)
	}
	klog.Infof("get %d datawarehouse", len(datawarehouse.Databases))
	for _, d := range datawarehouse.Databases {
		db := dwv1.FromApiDatabase(d)
		if apiDs, ok := existedDatawarehouses[GetDataWarehouseKey(&db)]; ok {
			same, err := dwv1.DeepCompareDataWarehouse(apiDs.Spec.DataWarehouse, &db)
			if err != nil {
				klog.Errorf("compare datasource %s err: %+v", db.Name, err)
			} else if !same {
				klog.V(4).Infof("need to update datawarehouse %s", db.Name)
				apiDs.Spec.DataWarehouse = &db
				apiDs.Status.UpdatedAt = metav1.Now()
				if apiDs.Spec.DataWarehouse.Tables == nil {
					apiDs.Spec.DataWarehouse.Tables = make([]dwv1.Table, 0)
				}
				for i := range apiDs.Spec.DataWarehouse.Tables {
					if apiDs.Spec.DataWarehouse.Tables[i].Properties == nil {
						apiDs.Spec.DataWarehouse.Tables[i].Properties = make([]dwv1.Property, 0)
					}
				}
				if err = r.Update(ctx, &apiDs); err != nil {
					klog.Errorf("cannot update datasource: %+v", err)
				}
			}
		} else {
			klog.V(4).Infof("need to create datawarehouse %s", GenerateName(&db))
			//klog.V(5).Infof("database=%+v", db)
			if db.Tables == nil {
				db.Tables = make([]dwv1.Table, 0)
			}
			for i, t := range db.Tables {
				if t.Properties == nil {
					db.Tables[i].Properties = make([]dwv1.Property, 0)
				}
			}
			ds := &nlptv1.Datasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      names.NewID(),
					Namespace: defaultNamespace,
				},
				Spec: nlptv1.DatasourceSpec{
					Name:          GenerateName(&db),
					Type:          nlptv1.DataWarehouseType,
					DataWarehouse: &db,
				},
				Status: nlptv1.DatasourceStatus{
					Status:    nlptv1.Normal,
					CreatedAt: metav1.Now(),
					UpdatedAt: metav1.Unix(0, 0),
				},
			}
			if err = r.Create(ctx, ds); err != nil {
				klog.Errorf("cannot create datasource: %+v", err)
			}
		}
		delete(existedDatawarehouses, GetDataWarehouseKey(&db))
	}
	for _, v := range existedDatawarehouses { //遍历本地资源 在最新资源中筛选已删除资源
		klog.V(4).Infof("need to delete datawarehouse %s", v.Spec.DataWarehouse.Name)
		if err = r.Delete(ctx, &v); err != nil {
			klog.Errorf("cannot delete datasource: %+v", err)
		}
	}
	return nil
}

func (r *DatasourceSynchronizer) Start(stop <-chan struct{}) error {
	// wait for cache up
	time.Sleep(time.Second * 10)
	wait.Until(func() {
		if err := r.SyncDatasources(); err != nil {
			klog.Errorf("sync datasource error: %+v", err)
		}
		// do not use wait.NerverStop
	}, time.Second*60, stop)
	return nil
}

// Important
func (*DatasourceSynchronizer) NeedLeaderElection() bool {
	return true
}
