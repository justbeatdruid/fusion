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
	AuthEnable     bool
	SuperUserToken string
	TokenSecret    string
}

func DefaultTopicOptions() *TopicOptions {
	return &TopicOptions{
		Host:           "10.160.32.24",
		Port:           30003,
		AuthEnable:     true,
		SuperUserToken: "/data/superUserToken",
		TokenSecret:    "/data/tokenSecret",
	}
}

func (o *TopicOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.Host, "pulsar-host", o.Host, "connect pulsar client")
	fs.IntVar(&o.Port, "pulsar-port", o.Port, "connect pulsar client")
	fs.BoolVar(&o.AuthEnable, "pulsar-auth-enable", o.AuthEnable, "enable pulsar authentication")
	fs.StringVar(&o.SuperUserToken, "pulsar-admin-token", o.SuperUserToken, "admin token of pulsar")
	fs.StringVar(&o.TokenSecret, "pulsar-token-secret", o.TokenSecret, "token secret file path")
}

func (o *TopicOptions) ParsePulsarSecret() error {
	if len(o.TokenSecret) == 0 {
		return fmt.Errorf("token secret path not set")
	}

	b := make([]byte, 5)
	b, err := ioutil.ReadFile(o.TokenSecret)
	if err != nil {
		return fmt.Errorf("cannot read token secret")
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
		return fmt.Errorf("cannot read superuser token")
	}
	o.SuperUserToken = string(b2)
	klog.Infof("ParsePulsarSecret: token: %+v, secret key: %+v", o.SuperUserToken, o.TokenSecret)
	return nil

}
