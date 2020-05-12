package options

import (
	"github.com/spf13/pflag"
)

type EtcdOptions struct {
	Endpoints string
	Timeout   int
	CertFile  string
	KeyFile   string
	CAFile    string
}

func DefaultEtcdOptions() *EtcdOptions {
	return &EtcdOptions{
		Endpoints: "http://127.0.0.1:2379",
		Timeout:   3,
	}
}

func (o *EtcdOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Endpoints, "etcd-endpoints", o.Endpoints, "etcd endpoints, split by comma \",\"")
	fs.IntVar(&o.Timeout, "etcd-timeout", o.Timeout, "etcd connection timeout in seconds")
	fs.StringVar(&o.CertFile, "etcd-cert-file", o.CertFile, "path of etcd cert file if use tls")
	fs.StringVar(&o.KeyFile, "etcd-key-file", o.KeyFile, "path of etcd key file if use tls")
	fs.StringVar(&o.CAFile, "etcd-ca-file", o.CAFile, "path of etcd ca file if use tls")
}
