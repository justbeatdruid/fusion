package options

import (
	"github.com/spf13/pflag"
)

type TopicOptions struct {
	Host       string
	Port       int
	AuthEnable bool
	AdminToken string
}

func DefaultTopicOptions() *TopicOptions {
	return &TopicOptions{
		Host:       "10.160.32.24",
		Port:       30003,
		AuthEnable: false,
		AdminToken: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhZG1pbiJ9.eNEbqeuUXxM7bsnP8gnxYq7hRkP50Rqc0nsWFRp8z6A",
	}
}

func (o *TopicOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Host, "pulsar-host", o.Host, "connect pulsar client")
	fs.IntVar(&o.Port, "pulsar-port", o.Port, "connect pulsar client")
	fs.BoolVar(&o.AuthEnable, "pulsar-auth-enable", o.AuthEnable, "enable pulsar authentication")
	fs.StringVar(&o.AdminToken, "pulsar-admin-token", o.AdminToken, "admin token of pulsar")
}
