package options

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	dynamicclient "k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"

	appconfig "github.com/chinamobile/nlpt/cmd/apiserver/app/config"
)

const userAgent = "nlpt"

type ServerRunOptions struct {
	ListenAddress string
	Master        string
	Kubeconfig    string
	CrdNamespace  string

	ConfigPath  string
	LocalConfig struct{}

	Datasource *DatasourceOptions
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		ListenAddress: ":8001",
		CrdNamespace:  os.Getenv("MY_POD_NAMESPACE"),

		Datasource: DefaultDatasourceOptions(),
	}
	if len(s.CrdNamespace) == 0 {
		klog.Infof("cannot find environmnent MY_POD_NAMESPACE, use default")
		s.CrdNamespace = "default"
	}
	return s
}

func (s *ServerRunOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	s.Datasource.AddFlags(fss.FlagSet("data source"))
	kfset := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(kfset)
	fss.FlagSet("klog").AddGoFlagSet(kfset)
	fs := fss.FlagSet("misc")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringVar(&s.ListenAddress, "listen", s.ListenAddress, "Address of app manager server listening on")
	return fss
}

func (s *ServerRunOptions) ParseOptions() error {
	if len(s.ConfigPath) == 0 {
		klog.Infof("config path not found, use default")
		s.defaultConfig()
		return nil
	}
	configFile, err := os.Open(s.ConfigPath)
	if err != nil {
		return fmt.Errorf("error opening config file: %+v", err)
	}

	decoder := json.NewDecoder(configFile)
	if err = decoder.Decode(&s.LocalConfig); err != nil {
		return fmt.Errorf("error parsing config file: %+v", err)
	}
	return nil
}

func (s *ServerRunOptions) defaultConfig() error {
	return nil
}

func (s *ServerRunOptions) Config() (*appconfig.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.Kubeconfig)
	if err != nil {
		return nil, err
	}
	//kubeconfig.DisableCompression = true

	client, err := clientset.NewForConfig(restclient.AddUserAgent(kubeconfig, userAgent))
	if err != nil {
		return nil, err
	}

	dynClient, err := dynamicclient.NewForConfig(restclient.AddUserAgent(kubeconfig, userAgent))
	if err != nil {
		return nil, err
	}
	c := &appconfig.Config{
		Client:     client,
		Dynamic:    dynClient,
		Kubeconfig: kubeconfig,

		DatasourceConfig: appconfig.NewDatasourceConfig(s.Datasource.Supported),
	}
	return c, nil
}