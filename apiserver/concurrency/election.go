package concurrency

import (
	"context"
	"fmt"
	"time"

	"github.com/satori/go.uuid"

	v3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"

	"k8s.io/klog"
)

func run_example() {
	t := time.NewTicker(time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-t.C:
			fmt.Printf("leader tik tak\n")
		}
	}
}

type Elector interface {
	Campaign(string, func())
}

type etcdElector struct {
	timeout time.Duration
	client  *v3.Client
}

func NewEtcdElector(endpoints, cert, key, ca string, timeoutInSecond int) (Elector, error) {
	cli, err := newEtcdClient(endpoints, cert, key, ca, timeoutInSecond)
	if err != nil {
		return nil, err
	}
	return &etcdElector{
		client:  cli,
		timeout: time.Duration(timeoutInSecond) * time.Second,
	}, nil
}

func (e *etcdElector) Campaign(category string, f func()) {
	var err error
	var session *concurrency.Session
	for {
		//sessionctx, sessioncancel := context.WithTimeout(context.Background(), e.timeout)
		//session, err = concurrency.NewSession(e.client, concurrency.WithTTL(5), concurrency.WithContext(sessionctx))
		session, err = concurrency.NewSession(e.client)
		//sessioncancel()
		if err != nil {
			klog.Errorf("cannot create session: %+v", err)
			continue
		}
		break
	}
	defer session.Close()

	candidator := uuid.NewV4().String()
	elect := concurrency.NewElection(session, "/nlpt-fusion-election/"+category)
	stop := make(chan struct{}, 0)
	go func() {
		cctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				currentLeader := string((<-elect.Observe(cctx)).Kvs[0].Value)
				klog.V(5).Infof("lock is now held by %s", currentLeader)
			}
		}
	}()

	klog.V(5).Infof("elector %s waiting\n", candidator)
	for {
		ctx := context.Background()
		//ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
		if err := elect.Campaign(ctx, candidator); err != nil {
			klog.Errorf("elector campaign error: %+v", err)
			continue
		}
		//cancel()
		break
	}
	close(stop)

	klog.V(5).Infof("elector %s running\n", candidator)
	f()
	klog.V(5).Infof("elector %s finished, ready to lease lock\n", candidator)

	for {
		//rctx, rcancel := context.WithTimeout(context.Background(), e.timeout)
		rctx := context.Background()
		if err := elect.Resign(rctx); err != nil {
			klog.Errorf("elector resign error: %+v", err)
			continue
		}
		//rcancel()
		break
	}
	klog.V(5).Infof("elector %s done\n", candidator)
}
