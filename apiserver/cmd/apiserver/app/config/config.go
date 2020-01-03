package config

import (
	apiserver "k8s.io/apiserver/pkg/server"
	dynamicclient "k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type Config struct {
	SecureServing   *apiserver.SecureServingInfo
	InsecureServing *apiserver.DeprecatedInsecureServingInfo
	Authentication  apiserver.AuthenticationInfo
	Authorization   apiserver.AuthorizationInfo

	Client  *clientset.Clientset
	Dynamic dynamicclient.Interface

	Kubeconfig *restclient.Config
}

func (c *Config) GetDynamicClient() dynamicclient.Interface {
	return c.Dynamic
}
