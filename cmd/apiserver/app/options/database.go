package options

import (
	"github.com/spf13/pflag"
)

type DatabaseOptions struct {
	Enabled  bool
	Type     string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Schema   string
}

func DefaultDatabaseOptions() *DatabaseOptions {
	return &DatabaseOptions{
		Enabled:  true,
		Type:     "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "123456",
		Database: "fusion",
	}
}

func (o *DatabaseOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.BoolVar(&o.Enabled, "database-enabled", o.Enabled, "store resources in database from cache, then query from database for listing")
	fs.StringVar(&o.Type, "database-type", o.Type, "database type")
	fs.StringVar(&o.Host, "database-host", o.Host, "database host")
	fs.IntVar(&o.Port, "database-port", o.Port, "database port")
	fs.StringVar(&o.Username, "database-username", o.Username, "database username")
	fs.StringVar(&o.Password, "database-password", o.Password, "database password")
	fs.StringVar(&o.Database, "database-databasename", o.Database, "database name")
	fs.StringVar(&o.Schema, "database-schema", o.Schema, "database schema (for postgres)")
}
