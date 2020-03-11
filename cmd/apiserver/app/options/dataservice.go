package options

import (
	"github.com/spf13/pflag"
)

type DataserviceOptions struct {
	MetadataHost string
	MetadataPort int
	DataHost     string
	DataPort     int
}

func DefaultDataserviceOptions() *DataserviceOptions {
	return &DataserviceOptions{
		MetadataHost: "127.0.0.1",
		MetadataPort: 27778,
		DataHost:     "127.0.0.1",
		DataPort:     27773,
	}
}

func (o *DataserviceOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.MetadataHost, "dataservice-metadata-host", o.MetadataHost, "host address where data service listen on with data provider in data platform backend")
	fs.IntVar(&o.MetadataPort, "dataservice-metadata-port", o.MetadataPort, "port of data provider in data platform backend")
	fs.StringVar(&o.DataHost, "dataservice-data-host", o.DataHost, "host address where data service listen on with data provider in data platform backend")
	fs.IntVar(&o.DataPort, "dataservice-data-port", o.DataPort, "port of data provider in data platform backend")
}
