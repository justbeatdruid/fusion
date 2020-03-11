package options

import (
	"github.com/spf13/pflag"
)

type CasOptions struct {
	Host string
	Port int
}

func DefaultCasOptions() *CasOptions {
	return &CasOptions{
		Host: "119.3.188.180",
		Port: 8000,
	}
}

func (o *CasOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Host, "cas-host", o.Host, "host of cas server which for quering user information")
	fs.IntVar(&o.Port, "cas-port", o.Port, "port of cas server which for quering user information")
}
