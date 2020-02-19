package options

import (
	"github.com/spf13/pflag"
)

type DataserviceOptions struct {
	Host string
	Port int
}

func DefaultDataserviceOptions() *DataserviceOptions {
	return &DataserviceOptions{
		Host: "127.0.0.1",
		Port: 27778,
	}
}

func (o *DataserviceOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Host, "dataservice-host", o.Host, "host address where data service listen on with data provider in data platform backend")
	fs.IntVar(&o.Port, "dataservice-port", o.Port, "port of data provider in data platform backend")
}
