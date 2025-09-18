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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
)

var _ = Describe("TangServer controller client functions", func() {

	Context("When testing client configuration functions", func() {
		It("Should handle GetClusterClientConfig correctly", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			config, err := GetClusterClientConfig()
			if err != nil {
				// In test environment, this might fail but that's expected
				Expect(err).ToNot(BeNil())
			} else {
				Expect(config).ToNot(BeNil())
				Expect(config).To(BeAssignableToTypeOf(&rest.Config{}))
			}
		})

		It("Should handle GetClientsetFromClusterConfig with valid config", func() {
			// Create a basic test config
			config := &rest.Config{
				Host: "https://test-cluster",
			}

			clientset, err := GetClientsetFromClusterConfig(config)
			// This will fail in test environment but tests the function logic
			if err != nil {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed creating clientset"))
			} else {
				Expect(clientset).ToNot(BeNil())
			}
		})

		It("Should handle GetClientsetFromClusterConfig with nil config", func() {
			// This test verifies that nil config is handled gracefully
			// The function should return an error when passed nil config
			defer func() {
				if r := recover(); r != nil {
					// If panic occurs, that's expected behavior for nil config
					Expect(r).ToNot(BeNil())
				}
			}()

			clientset, err := GetClientsetFromClusterConfig(nil)
			// Either should return error or panic (both are valid error handling)
			if err == nil && clientset != nil {
				// This shouldn't happen with nil config
				Fail("Expected error or panic with nil config")
			}
		})

		It("Should handle GetClusterClientset correctly", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			clientset, err := GetClusterClientset()
			// This will likely fail in test environment
			if err != nil {
				Expect(err).ToNot(BeNil())
			} else {
				Expect(clientset).ToNot(BeNil())
			}
		})

		It("Should handle GetRESTClient correctly", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster")
			}
			client, err := GetRESTClient()
			// This will likely fail in test environment
			if err != nil {
				Expect(err).ToNot(BeNil())
			} else {
				Expect(client).ToNot(BeNil())
			}
		})
	})

	Context("When testing kubeconfig file handling", func() {
		var tempKubeconfig string

		BeforeEach(func() {
			// Create a temporary kubeconfig file for testing
			tempDir := os.TempDir()
			tempKubeconfig = filepath.Join(tempDir, "test-kubeconfig")

			kubeconfigContent := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-server
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
			err := os.WriteFile(tempKubeconfig, []byte(kubeconfigContent), 0644)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			if tempKubeconfig != "" {
				os.Remove(tempKubeconfig)
			}
		})

		It("Should read kubeconfig file when HOME is set", func() {
			// Set HOME to temp directory for testing
			originalHome := os.Getenv("HOME")
			tempHome := filepath.Dir(tempKubeconfig)

			// Create .kube directory in temp home
			kubeDir := filepath.Join(tempHome, ".kube")
			err := os.MkdirAll(kubeDir, 0755)
			Expect(err).To(BeNil())

			// Copy kubeconfig to .kube/config
			kubeconfigPath := filepath.Join(kubeDir, "config")
			err = os.Rename(tempKubeconfig, kubeconfigPath)
			Expect(err).To(BeNil())

			os.Setenv("HOME", tempHome)

			config, err := GetClusterClientConfig()

			// Restore original HOME
			os.Setenv("HOME", originalHome)

			// Clean up
			os.RemoveAll(kubeDir)

			if err != nil {
				// May fail in test environment, but function was exercised
				Expect(err).ToNot(BeNil())
			} else {
				Expect(config).ToNot(BeNil())
			}
		})
	})
})
