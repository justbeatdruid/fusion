package concurrency

import (
	"context"
	"fmt"
	"time"

	v3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"

	log "k8s.io/klog"
)

type Mutex interface {
	Lock(category string) (func() error, error)
}

type etcdMutex struct {
	timeout time.Duration
	client  *v3.Client
}

func NewEtcdMutex(endpoints, cert, key, ca string, timeoutInSecond int) (Mutex, error) {
	cli, err := newEtcdClient(endpoints, cert, key, ca, timeoutInSecond)
	if err != nil {
		return nil, err
	}
	return &etcdMutex{
		client:  cli,
		timeout: time.Duration(timeoutInSecond) * time.Second,
	}, nil
}

func (e *etcdMutex) Lock(category string) (func() error, error) {
	sessionctx, sessioncancel := context.WithTimeout(context.Background(), e.timeout)
	session, err := concurrency.NewSession(e.client, concurrency.WithTTL(5), concurrency.WithContext(sessionctx))
	sessioncancel()
	if err != nil {
		err = fmt.Errorf("cannot create session: %+v", err)
		return nil, err
	}
	mtx := concurrency.NewMutex(session, "/nlpt-fusion-lock/"+category)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	//TODO Get lock directly if last lock timeout
	err = mtx.Lock(ctx)
	cancel()
	if err != nil {
		err = fmt.Errorf("cannot create lock: %+v", err)
		return nil, err
	}
	log.V(5).Infof("acquired lock for %s", category)
	if err != nil {
		go func() {
			if session != nil {
				session.Close()
			}
		}()
		return nil, err
	}
	return func() error {
		defer func() {
			if session != nil {
				session.Close()
			}
		}()
		innerctx, innercancel := context.WithTimeout(context.Background(), e.timeout)
		innererr := mtx.Unlock(innerctx)
		innercancel()
		if innererr != nil {
			log.Errorf("cannot release lock: %+v", innererr)
			return fmt.Errorf("cannot release lock: %+v", innererr)
		}
		innercancel()
		return nil
	}, nil
}
