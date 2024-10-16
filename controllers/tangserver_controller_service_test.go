/*
Copyright 2021.

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
	"context"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("TangServer controller service", func() {

	// Define utility constants for object names
	const (
		TangServerName                  = "test-tangserver-service"
		TangServerNamespace             = "default"
		TangServerResourceVersion       = "1"
		TangServerTestServiceListenPort = 8090
		TangServerTestIp                = "1.2.3.4"
		TangServerTestHostname          = "mylocalhost"
	)

	Context("When Creating TangServer", func() {
		It("Should be created with default listen port", func() {
			By("By creating a new TangServer with empty listen port")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangServerName,
					Namespace: TangServerNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			SetLogInstance(log.FromContext(ctx))
			service := getService(tangServer)
			Expect(service, Not(nil))
			Expect(service.TypeMeta.Kind, DEFAULT_SERVICE_TYPE)
			Expect(service.ObjectMeta.Name, getDefaultName(tangServer))
			Expect(service.Spec.Ports[0].Port, DEFAULT_SERVICE_PORT)
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should be created with specific service listen port", func() {
			By("By creating a new TangServer with non empty listen port")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangServerName,
					Namespace: TangServerNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					ServiceListenPort: TangServerTestServiceListenPort,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			SetLogInstance(log.FromContext(ctx))
			service := getService(tangServer)
			Expect(service, Not(nil))
			Expect(service.TypeMeta.Kind, DEFAULT_SERVICE_TYPE)
			Expect(service.ObjectMeta.Name, getDefaultName(tangServer))
			Expect(service.Spec.Ports[0].Port, TangServerTestServiceListenPort)
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should return a correct service url and related", func() {
			By("By creating a new TangServer")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangServerName,
					Namespace: TangServerNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					ServiceListenPort: TangServerTestServiceListenPort,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			serviceURL := getServiceURL(tangServer)
			Expect(len(serviceURL) > 0)
			serviceIpURL := getServiceIpURL(tangServer, TangServerTestIp)
			Expect(len(serviceIpURL) > 0)
			Expect(strings.Contains(serviceIpURL, TangServerTestIp))
			loadBalancer := corev1.LoadBalancerIngress{
				Hostname: TangServerTestHostname,
			}
			serviceIpExternalServiceURL := getExternalServiceURL(tangServer, loadBalancer)
			Expect(strings.Contains(serviceIpExternalServiceURL, TangServerTestHostname))
			loadBalancer = corev1.LoadBalancerIngress{
				IP: TangServerTestIp,
			}
			serviceIpExternalServiceURL = getExternalServiceURL(tangServer, loadBalancer)
			Expect(strings.Contains(serviceIpExternalServiceURL, TangServerTestIp))
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
	})
})
