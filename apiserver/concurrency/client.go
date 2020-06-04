package concurrency

import (
	"fmt"
	"strings"
	"time"

	v3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"

	"k8s.io/klog"
)

func newEtcdClient(endpoints, cert, key, ca string, timeoutInSecond int) (cli *v3.Client, err error) {
	dialTimeout := 3 * time.Second
	secure := "insecure"
	config := v3.Config{
		Endpoints:            strings.Split(endpoints, ","),
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    dialTimeout,
		DialKeepAliveTimeout: dialTimeout,
	}
	if len(cert) > 0 && len(key) > 0 && len(ca) > 0 {
		secure = "secure"
		tlsInfo := transport.TLSInfo{
			CertFile:      cert,
			KeyFile:       key,
			TrustedCAFile: ca,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("cannot configure tls: %+v", err)
		}
		config.TLS = tlsConfig
	}
	cli, err = v3.New(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create etcd client: %+v", err)
	}
	klog.Infof("create new %s etcd client with endpoints %s", secure, endpoints)
	return
}
