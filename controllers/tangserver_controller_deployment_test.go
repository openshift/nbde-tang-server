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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TangServer controller deployment", func() {

	// Define utility constants for object names
	const (
		TangserverName                   = "test-tangserver-deployment"
		TangserverNamespace              = "default"
		TangserverResourceVersion        = "1"
		TangServerTestReplicaAmount      = 4
		TangServerTestPodListenPort      = 8081
		TangServerTestSecret             = "thisisaverysimplesecretname"
		TangServerPrivateVolumeClaim     = "test-pvc"
		TangServerTestResourceRequestCpu = "10m"
		TangServerTestResourceRequestMem = "10M"
		TangServerTestResourceLimitCpu   = "20m"
		TangServerTestResourceLimitMem   = "20M"
	)

	Context("When Creating TangServer", func() {
		It("Should be created with default replica amount", func() {
			By("By creating a new TangServer with empty replica amount")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			deployment := getDeployment(tangServer)
			Expect(deployment, Not(nil))
			Expect(deployment.TypeMeta.Kind, DEFAULT_DEPLOYMENT_TYPE)
			Expect(deployment.ObjectMeta.Name, getDefaultName(tangServer))
			Expect(deployment.Spec.Replicas, DEFAULT_REPLICA_AMOUNT)
			Expect(deployment.Spec.Template, Not(nil))
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should be created with specific replica amount", func() {
			By("By creating a new TangServer with non empty replicas")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					Replicas: TangServerTestReplicaAmount,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			deployment := getDeployment(tangServer)
			Expect(deployment, Not(nil))
			Expect(deployment.TypeMeta.Kind, DEFAULT_DEPLOYMENT_TYPE)
			Expect(deployment.ObjectMeta.Name, getDefaultName(tangServer))
			Expect(deployment.Spec.Replicas, TangServerTestReplicaAmount)
			Expect(deployment.Spec.Template, Not(nil))
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
		It("Should be created with listen port, secret and requests", func() {
			By("By creating a new TangServer with non empty listen port and secret")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					PodListenPort:         TangServerTestPodListenPort,
					Secret:                TangServerTestSecret,
					PersistentVolumeClaim: TangServerPrivateVolumeClaim,
					ResourcesRequest: daemonsv1alpha1.ResourcesRequest{
						Cpu:    TangServerTestResourceRequestCpu,
						Memory: TangServerTestResourceRequestMem,
					},
					ResourcesLimit: daemonsv1alpha1.ResourcesLimit{
						Cpu:    TangServerTestResourceLimitCpu,
						Memory: TangServerTestResourceLimitMem,
					},
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())
			deployment := getDeployment(tangServer)
			Expect(deployment, Not(nil))
			Expect(deployment.TypeMeta.Kind, DEFAULT_DEPLOYMENT_TYPE)
			Expect(deployment.ObjectMeta.Name, getDefaultName(tangServer))
			Expect(deployment.Spec.Replicas, TangServerTestReplicaAmount)
			Expect(getDeploymentReadyReplicas(deployment), deployment.Status.ReadyReplicas)
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort, TangServerTestPodListenPort)
			rcpu, _ := resource.ParseQuantity(TangServerTestResourceRequestCpu)
			rmem, _ := resource.ParseQuantity(TangServerTestResourceRequestMem)
			lcpu, _ := resource.ParseQuantity(TangServerTestResourceLimitCpu)
			lmem, _ := resource.ParseQuantity(TangServerTestResourceLimitMem)
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU], rcpu)
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory], rmem)
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU], lcpu)
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory], lmem)
			Expect(deployment.Spec.Template.Spec.ImagePullSecrets[0].Name, TangServerTestSecret)
			Expect(deployment.Spec.Template.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName, TangServerPrivateVolumeClaim)
			Expect(isDeploymentReady(deployment), false)
			deployment.Status.Replicas = TangServerTestReplicaAmount
			deployment.Status.ReadyReplicas = deployment.Status.Replicas
			Expect(isDeploymentReady(deployment), true)
			err := k8sClient.Delete(ctx, tangServer)
			Expect(err, nil)
		})
	})
})
