package kubernetes

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	//utilpointer "k8s.io/utils/pointer"
)

//EnsureNamespace ...
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

//GetAllNamespaces ...
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

//CreateCronJob ...
func CreateCronJob(client *clientset.Clientset, name, schedule, iamgeName, containerName, namespace string, cmd []string) error {
	jobsClient := client.BatchV1beta1().CronJobs(namespace)
	cronJob := &v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: v1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:            containerName,
									Image:           iamgeName,
									Command:         cmd,
									ImagePullPolicy: v1.PullIfNotPresent,
								},
							},
							RestartPolicy: v1.RestartPolicyNever,
						},
					},
				},
			},
			SuccessfulJobsHistoryLimit: new(int32),
		},
	}
	*cronJob.Spec.SuccessfulJobsHistoryLimit = 1
	result, err := jobsClient.Create(cronJob)
	klog.Infof("create job: result %+v, err:%v", result, err)
	if err != nil {
		return fmt.Errorf("create error err: %+v, result:%v", err, result)
	}
	return nil

}

//CreateJob ...
func CreateJob(client *clientset.Clientset, name, imageName, containerName, namespace string) error {
	jobsClient := client.BatchV1().Jobs(namespace)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  containerName,
							Image: imageName,
						},
					},
				},
			},
		},
	}
	result, err := jobsClient.Create(job)
	klog.V(5).Infof("create job: result %+v, err:%v", result, err)
	if err != nil {
		return fmt.Errorf("create error err: %+v, result:%v", err, result)
	}
	return nil
}

//DeleteCronJob ...
func DeleteCronJob(client *clientset.Clientset, name, namespace string) error {
	jobsClient := client.BatchV1beta1().CronJobs(namespace)
	return jobsClient.Delete(name, &metav1.DeleteOptions{})
}

//DeleteJob ...
func DeleteJob(client *clientset.Clientset, name, namespace string) error {
	jobsClient := client.BatchV1().Jobs(namespace)
	return jobsClient.Delete(name, &metav1.DeleteOptions{})
}
