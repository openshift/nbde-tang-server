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
// +kubebuilder:docs-gen:collapse=Apache License

package controllers

import (
	"context"
	"crypto/tls"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const FAKE_RECORDER_BUFFER = 1000

// getOptions returns fake options for local controller testing
func getOptions(scheme *runtime.Scheme) *ctrl.Options {
	metricsAddr := "localhost:7070"
	probeAddr := "localhost:7071"
	enableLeaderElection := false
	disableHTTP2 := func(c *tls.Config) {
		c.NextProtos = []string{"http/1.1"}
	}

	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: true,
		TLSOpts:       []func(*tls.Config){disableHTTP2},
	}

	return &ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e44fa0d3.redhat.com",
	}
}

// getClientOptions returns fake options for local controller testing
func getClientOptions(scheme *runtime.Scheme) *client.Options {
	return &client.Options{
		Scheme: scheme,
	}
}

// isCluster checks for environment variable value to run test on cluster
func isCluster() bool {
	return os.Getenv("CLUSTER_TANG_OPERATOR_TEST") == "1" || os.Getenv("CLUSTER_TANG_OPERATOR_TEST") == "y"
}

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("TangServer controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		TangserverName      = daemonsv1alpha1.DefaultTestName
		TangserverNameNoUID = daemonsv1alpha1.DefaultTestNameNoUID
		TangserverNamespace = daemonsv1alpha1.DefaultTestNamespace
	)

	Context("When Creating Simple TangServer", func() {
		It("Should be created with no error", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			By("By creating a new TangServer with default specs")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Succeed())

			By("By checking complete empty specs are valid")
			emptyTangServer := &daemonsv1alpha1.TangServer{}
			Expect(emptyTangServer.Spec.KeyPath).Should(Equal(""))
			Expect(emptyTangServer.Spec.Replicas).Should(Equal(uint32(0)))
			Expect(emptyTangServer.Spec.Image).Should(Equal(""))
			Expect(emptyTangServer.Spec.Version).Should(Equal(""))
		})
		It("Should not be created", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			By("By creating a TangServer that already exist")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "/",
				},
			}
			Expect(k8sClient.Create(ctx, tangServer)).Should(Not(Succeed()))
		})
		It("Reconcile should be executed with no error", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			By("By creating a new TangServer with default specs")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverNameNoUID,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "/",
				},
			}
			s := scheme.Scheme
			s.AddKnownTypes(daemonsv1alpha1.GroupVersion, tangServer)
			mgr, _ := ctrl.NewManager(ctrl.GetConfigOrDie(), *getOptions(s))
			nc, _ := client.New(ctrl.GetConfigOrDie(), *getClientOptions(s))
			rec := TangServerReconciler{
				Client:   nc,
				Scheme:   s,
				Recorder: record.NewFakeRecorder(FAKE_RECORDER_BUFFER),
			}
			err := rec.SetupWithManager(mgr)
			Expect(err, nil)
			_, err = rec.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: TangserverNamespace,
					Name:      TangserverNameNoUID,
				},
			})
			Expect(err, nil)
		})
		It("Double Reconcile should be executed with no error", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			By("By creating a new TangServer with default specs")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "/",
				},
			}
			s := scheme.Scheme
			s.AddKnownTypes(daemonsv1alpha1.GroupVersion, tangServer)
			mgr, _ := ctrl.NewManager(ctrl.GetConfigOrDie(), *getOptions(s))
			nc, _ := client.New(ctrl.GetConfigOrDie(), *getClientOptions(s))
			rec := TangServerReconciler{
				Client:   nc,
				Scheme:   s,
				Recorder: record.NewFakeRecorder(FAKE_RECORDER_BUFFER),
			}
			err := rec.SetupWithManager(mgr)
			Expect(err, nil)
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
			}
			_, err = rec.Reconcile(ctx, req)
			Expect(err, nil)
			_, err = rec.Reconcile(ctx, req)
			Expect(err, nil)
		})
		It("Reconcile with deletion time executed with no error", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			By("By creating a new TangServer with default specs")
			ctx := context.Background()
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "/",
				},
			}
			n := metav1.Now()
			tangServer.ObjectMeta.SetDeletionTimestamp(&n)
			s := scheme.Scheme
			s.AddKnownTypes(daemonsv1alpha1.GroupVersion, tangServer)
			mgr, _ := ctrl.NewManager(ctrl.GetConfigOrDie(), *getOptions(s))
			nc, _ := client.New(ctrl.GetConfigOrDie(), *getClientOptions(s))
			rec := TangServerReconciler{
				Client:   nc,
				Scheme:   s,
				Recorder: record.NewFakeRecorder(FAKE_RECORDER_BUFFER),
			}
			err := rec.SetupWithManager(mgr)
			Expect(err, nil)
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      TangserverName,
					Namespace: TangserverNamespace,
				},
			}
			_, err = rec.Reconcile(ctx, req)
			Expect(err, nil)
		})

	})
})
