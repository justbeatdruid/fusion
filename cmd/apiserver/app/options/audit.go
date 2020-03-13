package options

import (
	"github.com/spf13/pflag"
)

type AuditOptions struct {
	Host string
	Port int
}

func DefaultAuditOptions() *AuditOptions {
	return &AuditOptions{
		Host: "127.0.0.1",
		Port: 80,
	}
}

func (o *AuditOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Host, "audit-host", o.Host, "host address where audit server listen on")
	fs.IntVar(&o.Port, "audit-port", o.Port, "port of audit server")
}
