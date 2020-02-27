package options

import (
	"github.com/spf13/pflag"
)

type TopicOptions struct {
	Host string
	Port int
}

func DefaultTopicOptions() *TopicOptions {
	return &TopicOptions{
		Host: "127.0.0.1",
		Port: 30003,
	}
}

func (o *TopicOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Host, "pulsar-host", o.Host, "connect pulsar client")
	fs.IntVar(&o.Port, "pulsar-port", o.Port, "connect pulsar client")
}
