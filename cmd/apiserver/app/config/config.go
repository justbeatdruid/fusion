package config

import (
	apiserver "k8s.io/apiserver/pkg/server"
	dynamicclient "k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
)

type Config struct {
	SecureServing   *apiserver.SecureServingInfo
	InsecureServing *apiserver.DeprecatedInsecureServingInfo
	Authentication  apiserver.AuthenticationInfo
	Authorization   apiserver.AuthorizationInfo

	Client  *clientset.Clientset
	Dynamic dynamicclient.Interface

	Kubeconfig *restclient.Config

	DatasourceConfig *DatasourceConfig

	DataserviceConnector dw.Connector

	TopicConfig *TopicConfig
}

func (c *Config) GetDynamicClient() dynamicclient.Interface {
	return c.Dynamic
}
