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

	Describe("Service utility functions", func() {
		var tangServer *daemonsv1alpha1.TangServer

		BeforeEach(func() {
			tangServer = &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "test-namespace",
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					ServiceListenPort: 8080,
					ServiceType:       "LoadBalancer",
				},
			}
		})

		Context("getServiceName", func() {
			It("should return correct service name", func() {
				expected := "service-test-server"
				actual := getServiceName(tangServer)
				Expect(actual).To(Equal(expected))
			})
		})

		Context("getServicePort", func() {
			It("should return specified service port", func() {
				expected := int32(8080)
				actual := getServicePort(tangServer)
				Expect(actual).To(Equal(expected))
			})

			It("should return default port when not specified", func() {
				tangServer.Spec.ServiceListenPort = 0
				expected := int32(DEFAULT_SERVICE_PORT)
				actual := getServicePort(tangServer)
				Expect(actual).To(Equal(expected))
			})
		})

		Context("getServiceType", func() {
			It("should return LoadBalancer when specified", func() {
				tangServer.Spec.ServiceType = "LoadBalancer"
				actual := getServiceType(tangServer)
				expected := corev1.ServiceTypeLoadBalancer
				Expect(actual).To(Equal(expected))
			})

			It("should return ClusterIP when specified", func() {
				tangServer.Spec.ServiceType = "ClusterIP"
				actual := getServiceType(tangServer)
				expected := corev1.ServiceTypeClusterIP
				Expect(actual).To(Equal(expected))
			})

			It("should return NodePort when specified", func() {
				tangServer.Spec.ServiceType = "NodePort"
				actual := getServiceType(tangServer)
				expected := corev1.ServiceTypeNodePort
				Expect(actual).To(Equal(expected))
			})

			It("should return default LoadBalancer when empty", func() {
				tangServer.Spec.ServiceType = ""
				actual := getServiceType(tangServer)
				expected := corev1.ServiceTypeLoadBalancer
				Expect(actual).To(Equal(expected))
			})

			It("should return LoadBalancer for invalid service type", func() {
				tangServer.Spec.ServiceType = "InvalidType"
				actual := getServiceType(tangServer)
				expected := corev1.ServiceTypeLoadBalancer
				Expect(actual).To(Equal(expected))
			})
		})

		Context("service configuration", func() {
			It("should handle service configuration correctly", func() {
				tangServer.Spec.ServiceListenPort = 9000
				tangServer.Spec.ServiceType = "LoadBalancer"

				Expect(tangServer.Spec.ServiceListenPort).To(Equal(int32(9000)))
				Expect(tangServer.Spec.ServiceType).To(Equal("LoadBalancer"))
			})
		})
	})

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
