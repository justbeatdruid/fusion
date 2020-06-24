package options

import (
	"fmt"
	"github.com/spf13/pflag"
	"io/ioutil"
	"k8s.io/klog"
)

type TopicOptions struct {
	Host           string
	Port           int
	HttpPort       int
	AuthEnable     bool
	SuperUserToken string
	TokenSecret    string
	PrestoHost     string
	PrestoPort     int
}

func DefaultTopicOptions() *TopicOptions {
	return &TopicOptions{
		Host:           "10.160.32.24",
		Port:           30004,
		HttpPort:       30002,
		AuthEnable:     true,
		SuperUserToken: "/data/pulsar-secret/superUserToken",
		TokenSecret:    "/data/pulsar-secret/tokenSecret",
		PrestoHost:     "10.160.32.24",
		PrestoPort:     30004,
	}
}

func (o *TopicOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Host, "pulsar-host", o.Host, "connect pulsar client")
	fs.IntVar(&o.Port, "pulsar-port", o.Port, "connect pulsar client")
	fs.IntVar(&o.HttpPort, "pulsar-http-port", o.Port, "connect pulsar http rest api")
	fs.BoolVar(&o.AuthEnable, "pulsar-auth-enable", o.AuthEnable, "enable pulsar authentication")
	fs.StringVar(&o.SuperUserToken, "pulsar-admin-token", o.SuperUserToken, "admin token of pulsar")
	fs.StringVar(&o.TokenSecret, "pulsar-token-secret", o.TokenSecret, "token secret file path")
	fs.StringVar(&o.PrestoHost, "presto-host", o.PrestoHost, "connect pulsar presto server")
	fs.IntVar(&o.PrestoPort, "presto-port", o.PrestoPort, "connect pulsar presto server")

}

func (o *TopicOptions) ParsePulsarSecret() error {
	if len(o.TokenSecret) == 0 {
		return fmt.Errorf("token secret path not set")
	}

	b := make([]byte, 5)
	b, err := ioutil.ReadFile(o.TokenSecret)
	if err != nil {
		return fmt.Errorf("cannot read token secret: %+v", err)
	}
	o.TokenSecret = string(b)

	if len(o.TokenSecret) == 0 {
		return fmt.Errorf("token secret path not set")
	}

	if len(o.SuperUserToken) == 0 {
		return fmt.Errorf("SuperUserToken path not set")
	}
	b2 := make([]byte, 5)
	b2, err = ioutil.ReadFile(o.SuperUserToken)
	if err != nil {
		return fmt.Errorf("cannot read superuser token: %+v", err)
	}
	o.SuperUserToken = string(b2)
	klog.Infof("ParsePulsarSecret: token: %+v, secret key: %+v", o.SuperUserToken, o.TokenSecret)
	return nil

}
