package cache

import (
	"k8s.io/client-go/tools/cache"

	"github.com/chinamobile/nlpt/apiserver/database"

	"k8s.io/klog"
)

var gvm = make(map[string]cache.ResourceEventHandlerFuncs)

func initGVM(db *database.DatabaseConnection) {
	for _, g := range gvs {
		var addFunc, deleteFunc func(interface{}) error
		var updateFunc func(o, n interface{}) error
		switch g.resource {
		case "applications":
			addFunc = db.AddApplication
			updateFunc = db.UpdateApplication
			deleteFunc = db.DeleteApplication
		default:
			goto next
		}
		gvm[g.resource] = cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if err := addFunc(obj); err != nil {
					klog.Errorf("add func error: %+v", err)
				}
			},
			UpdateFunc: func(oldobj, newobj interface{}) {
				if err := updateFunc(oldobj, newobj); err != nil {
					klog.Errorf("update func error: %+v", err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if err := deleteFunc(obj); err != nil {
					klog.Errorf("delete func error: %+v", err)
				}
			},
		}
	next:
	}
}
