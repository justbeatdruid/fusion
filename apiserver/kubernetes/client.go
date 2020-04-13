package kubernetes

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func EnsureNamespace(client *clientset.Clientset, namespace string) error {
	_, err := client.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("cannot get namespace: %+v", err)
	}
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err = client.CoreV1().Namespaces().Create(ns)
	if err != nil {
		return fmt.Errorf("cannot create namespace: %+v", err)
	}
	return nil
}

func GetAllNamespaces(client *clientset.Clientset) ([]string, error) {
	nsl, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot list namespace: %+v", err)
	}
	namespaces := make([]string, len(nsl.Items))
	for i := range nsl.Items {
		namespaces[i] = nsl.Items[i].ObjectMeta.Name
	}
	return namespaces, nil
}
