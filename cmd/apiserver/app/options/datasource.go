package options

import (
	"github.com/spf13/pflag"
)

type DatasourceOptions struct {
	Supported string
}

func DefaultDatasourceOptions() *DatasourceOptions {
	return &DatasourceOptions{
		Supported: "*",
	}
}

func (o *DatasourceOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Supported, "supported-datasource-type", o.Supported, "supported datasource type, devided by comma. for example, \"mysql,postgres,hive\". \"*\" for no limit")
}
