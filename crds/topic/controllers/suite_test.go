/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	nlptv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var op *Connector

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

func TestOperator_CreateTopic(t *testing.T) {
	var test = "hello"
	fmt.Println(len(test))

	test = "你好"
	fmt.Println(len(test))

	op = new(Connector)
	op.Host = "10.160.32.24"
	op.Port = 30002
	op.AuthEnable = true
	op.SuperUserToken = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhZG1pbiJ9.eNEbqeuUXxM7bsnP8gnxYq7hRkP50Rqc0nsWFRp8z6A"
	topic := new(nlptv1.Topic)
	topic.Spec.TopicGroup = "default"
	topic.Spec.Name = "nonPartitionTopic"

	op.CreateTopic(topic)
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = nlptv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

func TestConnector_GetStats(t *testing.T) {
	var tmp float64 = 0
	fmt.Println(fmt.Sprintf("%.8f", tmp))
	var tmp2 float64 = 5.000
	fmt.Println(tmp2)
	tmpValue, _ := strconv.ParseFloat(fmt.Sprintf("%.8f", tmp), 64)
	fmt.Println(tmpValue)
	op = new(Connector)
	op.Host = "10.160.32.24"
	op.Port = 30002
	op.AuthEnable = true
	op.SuperUserToken = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhZG1pbiJ9.eNEbqeuUXxM7bsnP8gnxYq7hRkP50Rqc0nsWFRp8z6A"
	topic := nlptv1.Topic{}
	topic.Spec.TopicGroup = "default"
	topic.Spec.Name = "nonPartitionTopic"
	topic.Spec.PartitionNum = 1
	topic.Spec.Partitioned = true

	op.GetStats(topic)
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
