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

package v1alpha1

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestV1Alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V1Alpha1 Suite")
}

var _ = Describe("TangServer Types", func() {
	Describe("TangServerSpec", func() {
		var spec TangServerSpec

		BeforeEach(func() {
			spec = TangServerSpec{
				Replicas:               3,
				KeyPath:                "/var/db/tang",
				Image:                  "registry.redhat.io/rhel8/tang",
				Version:                "latest",
				ServiceListenPort:      7500,
				PodListenPort:          7500,
				RequiredActiveKeyPairs: 2,
				KeyRefreshInterval:     300,
				ServiceType:            "LoadBalancer",
				PersistentVolumeClaim:  "tang-pvc",
				HealthScript:           "/usr/bin/tangd-healthcheck",
				Secret:                 "tang-secret",
			}
		})

		It("should have correct field values", func() {
			Expect(spec.Replicas).To(Equal(int32(3)))
			Expect(spec.KeyPath).To(Equal("/var/db/tang"))
			Expect(spec.Image).To(Equal("registry.redhat.io/rhel8/tang"))
			Expect(spec.Version).To(Equal("latest"))
			Expect(spec.ServiceListenPort).To(Equal(int32(7500)))
			Expect(spec.PodListenPort).To(Equal(int32(7500)))
			Expect(spec.RequiredActiveKeyPairs).To(Equal(uint32(2)))
			Expect(spec.KeyRefreshInterval).To(Equal(uint32(300)))
			Expect(spec.ServiceType).To(Equal("LoadBalancer"))
			Expect(spec.PersistentVolumeClaim).To(Equal("tang-pvc"))
			Expect(spec.HealthScript).To(Equal("/usr/bin/tangd-healthcheck"))
			Expect(spec.Secret).To(Equal("tang-secret"))
		})

		It("should support resource requests and limits", func() {
			spec.ResourcesRequest = ResourcesRequest{
				Cpu:    "100m",
				Memory: "128Mi",
			}
			spec.ResourcesLimit = ResourcesLimit{
				Cpu:    "500m",
				Memory: "512Mi",
			}

			Expect(spec.ResourcesRequest.Cpu).To(Equal("100m"))
			Expect(spec.ResourcesRequest.Memory).To(Equal("128Mi"))
			Expect(spec.ResourcesLimit.Cpu).To(Equal("500m"))
			Expect(spec.ResourcesLimit.Memory).To(Equal("512Mi"))
		})

		It("should support hidden keys configuration", func() {
			spec.HiddenKeys = []TangServerHiddenKeys{
				{Sha1: "test-sha1-key"},
				{Sha256: "test-sha256-key"},
			}

			Expect(spec.HiddenKeys).To(HaveLen(2))
			Expect(spec.HiddenKeys[0].Sha1).To(Equal("test-sha1-key"))
			Expect(spec.HiddenKeys[1].Sha256).To(Equal("test-sha256-key"))
		})
	})

	Describe("TangServerStatus", func() {
		var status TangServerStatus

		BeforeEach(func() {
			status = TangServerStatus{
				Ready:              int32(2),
				Running:            int32(2),
				ServiceExternalURL: "http://tang-service:7500",
			}
		})

		It("should have correct status fields", func() {
			Expect(status.Ready).To(Equal(int32(2)))
			Expect(status.Running).To(Equal(int32(2)))
			Expect(status.ServiceExternalURL).To(Equal("http://tang-service:7500"))
		})

		It("should support empty status", func() {
			emptyStatus := TangServerStatus{}
			Expect(emptyStatus.Ready).To(Equal(int32(0)))
			Expect(emptyStatus.Running).To(Equal(int32(0)))
			Expect(emptyStatus.ServiceExternalURL).To(Equal(""))
		})
	})

	Describe("TangServer", func() {
		var tangServer TangServer

		BeforeEach(func() {
			tangServer = TangServer{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "nbde.openshift.io/v1alpha1",
					Kind:       "TangServer",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tangserver",
					Namespace: "test-namespace",
				},
				Spec: TangServerSpec{
					Replicas:          2,
					KeyPath:           "/var/db/tang",
					ServiceListenPort: 7500,
				},
				Status: TangServerStatus{
					Ready:              int32(1),
					ServiceExternalURL: "http://test-service:7500",
				},
			}
		})

		It("should have correct metadata", func() {
			Expect(tangServer.APIVersion).To(Equal("nbde.openshift.io/v1alpha1"))
			Expect(tangServer.Kind).To(Equal("TangServer"))
			Expect(tangServer.Name).To(Equal("test-tangserver"))
			Expect(tangServer.Namespace).To(Equal("test-namespace"))
		})

		It("should have correct spec and status", func() {
			Expect(tangServer.Spec.Replicas).To(Equal(int32(2)))
			Expect(tangServer.Spec.KeyPath).To(Equal("/var/db/tang"))
			Expect(tangServer.Spec.ServiceListenPort).To(Equal(int32(7500)))

			Expect(tangServer.Status.Ready).To(Equal(int32(1)))
			Expect(tangServer.Status.ServiceExternalURL).To(Equal("http://test-service:7500"))
		})
	})

	Describe("TangServerList", func() {
		It("should support list operations", func() {
			tangServerList := TangServerList{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "nbde.openshift.io/v1alpha1",
					Kind:       "TangServerList",
				},
				Items: []TangServer{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "tangserver-1",
						},
						Spec: TangServerSpec{
							Replicas: int32(1),
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "tangserver-2",
						},
						Spec: TangServerSpec{
							Replicas: int32(2),
						},
					},
				},
			}

			Expect(tangServerList.Items).To(HaveLen(2))
			Expect(tangServerList.Items[0].Name).To(Equal("tangserver-1"))
			Expect(tangServerList.Items[1].Name).To(Equal("tangserver-2"))
			Expect(tangServerList.Items[0].Spec.Replicas).To(Equal(int32(1)))
			Expect(tangServerList.Items[1].Spec.Replicas).To(Equal(int32(2)))
		})
	})

	Describe("Resource types", func() {
		It("should support ResourcesRequest", func() {
			req := ResourcesRequest{
				Cpu:    "100m",
				Memory: "128Mi",
			}

			Expect(req.Cpu).To(Equal("100m"))
			Expect(req.Memory).To(Equal("128Mi"))
		})

		It("should support ResourcesLimit", func() {
			limit := ResourcesLimit{
				Cpu:    "1000m",
				Memory: "1Gi",
			}

			Expect(limit.Cpu).To(Equal("1000m"))
			Expect(limit.Memory).To(Equal("1Gi"))
		})

		It("should support TangServerHiddenKeys", func() {
			hiddenKey := TangServerHiddenKeys{
				Sha1:   "test-sha1",
				Sha256: "test-sha256",
			}

			Expect(hiddenKey.Sha1).To(Equal("test-sha1"))
			Expect(hiddenKey.Sha256).To(Equal("test-sha256"))
		})
	})

	Describe("DeepCopy operations", func() {
		var original *TangServer

		BeforeEach(func() {
			original = &TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "original-server",
					Namespace: "test-namespace",
				},
				Spec: TangServerSpec{
					Replicas: 3,
					KeyPath:  "/var/db/tang",
					Image:    "registry.redhat.io/rhel8/tang",
					Version:  "latest",
					HiddenKeys: []TangServerHiddenKeys{
						{Sha1: "test-key"},
					},
				},
				Status: TangServerStatus{
					Ready:              int32(2),
					ServiceExternalURL: "http://test:7500",
				},
			}
		})

		It("should create deep copy correctly", func() {
			copied := original.DeepCopy()

			Expect(copied).NotTo(BeIdenticalTo(original))
			Expect(copied.Name).To(Equal(original.Name))
			Expect(copied.Namespace).To(Equal(original.Namespace))
			Expect(copied.Spec.Replicas).To(Equal(original.Spec.Replicas))
			Expect(copied.Spec.KeyPath).To(Equal(original.Spec.KeyPath))
			Expect(copied.Spec.HiddenKeys).To(HaveLen(1))
			Expect(copied.Status.Ready).To(Equal(int32(2)))

			// Modify the copy to ensure independence
			copied.Name = "modified-server"
			copied.Spec.Replicas = 5
			copied.Spec.HiddenKeys[0].Sha1 = "modified-key"

			Expect(original.Name).To(Equal("original-server"))
			Expect(original.Spec.Replicas).To(Equal(int32(3)))
			Expect(original.Spec.HiddenKeys[0].Sha1).To(Equal("test-key"))
		})

		It("should create deep copy object correctly", func() {
			copied := original.DeepCopyObject()
			tangServerCopy, ok := copied.(*TangServer)

			Expect(ok).To(BeTrue())
			Expect(tangServerCopy).NotTo(BeIdenticalTo(original))
			Expect(tangServerCopy.Name).To(Equal(original.Name))
		})
	})

	Describe("Runtime object behavior", func() {
		It("should implement runtime.Object interface", func() {
			tangServer := &TangServer{}
			var obj runtime.Object = tangServer

			Expect(obj).NotTo(BeNil())
			Expect(obj.GetObjectKind()).NotTo(BeNil())
		})

		It("should handle TangServerList as runtime object", func() {
			tangServerList := &TangServerList{}
			var obj runtime.Object = tangServerList

			Expect(obj).NotTo(BeNil())
			Expect(obj.GetObjectKind()).NotTo(BeNil())
		})
	})
})
