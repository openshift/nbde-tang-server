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
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TangServer controller key status functions", func() {

	Context("When testing key status file path functions", func() {
		It("Should return correct key status file name", func() {
			result := keyStatusFile()
			Expect(result).To(Equal("key_status.txt"))
		})

		It("Should return correct key status file path", func() {
			testTangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tang",
					Namespace: "test-namespace",
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "/test/path",
				},
			}
			keyAssocInfo := KeyAssociationInfo{
				KeyInfo: &KeyObtainInfo{
					TangServer: testTangServer,
				},
			}
			result := keyStatusFilePath(keyAssocInfo)
			expected := "/test/path" + "/" + keyStatusFile()
			Expect(result).To(Equal(expected))
		})

		It("Should return correct key status lock file path", func() {
			testTangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tang",
					Namespace: "test-namespace",
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "/test/path",
				},
			}
			keyAssocInfo := KeyAssociationInfo{
				KeyInfo: &KeyObtainInfo{
					TangServer: testTangServer,
				},
			}
			result := keyStatusLockFilePath(keyAssocInfo)
			expected := "/test/path" + "/" + keyStatusFile() + ".lock"
			Expect(result).To(Equal(expected))
		})

		It("Should return correct key status file path with TangServer", func() {
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tang",
					Namespace: "test-namespace",
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "/custom/path",
				},
			}

			result := keyStatusFilePathWithTangServer(tangServer)
			expected := "/custom/path/" + keyStatusFile()
			Expect(result).To(Equal(expected))
		})

		It("Should use default path when TangServer KeyPath is empty", func() {
			tangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tang",
					Namespace: "test-namespace",
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: "",
				},
			}

			result := keyStatusFilePathWithTangServer(tangServer)
			expected := DEFAULT_DEPLOYMENT_KEY_PATH + "/" + keyStatusFile()
			Expect(result).To(Equal(expected))
		})
	})

	Context("When testing key association functions", func() {
		var tempDir string
		var keyInfo KeyObtainInfo
		var tangServer *daemonsv1alpha1.TangServer

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "tang-test-*")
			Expect(err).To(BeNil())

			tangServer = &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tang",
					Namespace: "test-namespace",
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: tempDir,
				},
			}

			keyInfo = KeyObtainInfo{
				PodName:    "test-pod",
				Namespace:  "test-namespace",
				DbPath:     tempDir,
				TangServer: tangServer,
			}
		})

		AfterEach(func() {
			if tempDir != "" {
				os.RemoveAll(tempDir)
			}
		})

		It("Should test key association data structure", func() {
			sha1 := "test-sha1"
			sha256 := "test-sha256"
			signingKey := "signing-key-path"
			encryptionKey := "encryption-key-path"

			keyAssoc := KeyAssociation{
				Sha1:          sha1,
				Sha256:        sha256,
				SigningKey:    signingKey,
				EncriptionKey: encryptionKey,
			}

			Expect(keyAssoc.Sha1).To(Equal(sha1))
			Expect(keyAssoc.Sha256).To(Equal(sha256))
			Expect(keyAssoc.SigningKey).To(Equal(signingKey))
			Expect(keyAssoc.EncriptionKey).To(Equal(encryptionKey))

			// Test JSON marshaling
			jsonData, err := json.Marshal(keyAssoc)
			Expect(err).To(BeNil())
			Expect(jsonData).ToNot(BeEmpty())

			var parsedAssoc KeyAssociation
			err = json.Unmarshal(jsonData, &parsedAssoc)
			Expect(err).To(BeNil())
			Expect(parsedAssoc.SigningKey).To(Equal(signingKey))
			Expect(parsedAssoc.EncriptionKey).To(Equal(encryptionKey))
		})

		It("Should handle empty key association", func() {
			keyAssoc := KeyAssociation{}
			jsonData, err := json.Marshal(keyAssoc)
			Expect(err).To(BeNil())
			Expect(jsonData).ToNot(BeEmpty())

			var parsedAssoc KeyAssociation
			err = json.Unmarshal(jsonData, &parsedAssoc)
			Expect(err).To(BeNil())
			Expect(parsedAssoc.Sha1).To(Equal(""))
			Expect(parsedAssoc.Sha256).To(Equal(""))
		})

		It("Should test dumpKeyAssociation function signature", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster/pod execution")
			}
			// Test that the function exists and can be called
			// Note: This function likely requires pod execution which we can't fully test here
			keyAssocInfo := KeyAssociationInfo{
				KeyInfo:  &keyInfo,
				KeyAssoc: KeyAssociation{},
			}

			// Just verify the function can be called without panicking
			// The actual implementation requires pod command execution
			Expect(func() {
				_ = dumpKeyAssociation(keyAssocInfo)
			}).ToNot(Panic())
		})
	})

	Context("When testing key status file operations", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "tang-keystatus-test-*")
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			if tempDir != "" {
				os.RemoveAll(tempDir)
			}
		})

		It("Should generate correct status file path", func() {
			testTangServer := &daemonsv1alpha1.TangServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tang",
					Namespace: "test-namespace",
				},
				Spec: daemonsv1alpha1.TangServerSpec{
					KeyPath: tempDir,
				},
			}
			keyAssocInfo := KeyAssociationInfo{
				KeyInfo: &KeyObtainInfo{
					TangServer: testTangServer,
				},
			}
			statusPath := keyStatusFilePath(keyAssocInfo)
			expectedPath := filepath.Join(tempDir, keyStatusFile())
			Expect(statusPath).To(Equal(expectedPath))
		})

		It("Should test writeStatusFile function signature", func() {
			if !isCluster() {
				Skip("Avoiding test that requires cluster/pod execution")
			}
			// Test that the function exists and can be called
			sha1 := "test-sha1"
			sha256 := "test-sha256"
			signing := "signing-key"
			encryption := "encryption-key"

			// Create a proper keyInfo structure for the test
			testKeyInfo := KeyObtainInfo{
				PodName:   "test-pod",
				Namespace: "test-namespace",
				DbPath:    tempDir,
				TangServer: &daemonsv1alpha1.TangServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-tang",
						Namespace: "test-namespace",
					},
					Spec: daemonsv1alpha1.TangServerSpec{
						KeyPath: tempDir,
					},
				},
			}

			// Note: This function likely requires pod execution which we can't fully test here
			// Just verify the function can be called without panicking
			Expect(func() {
				_ = writeStatusFile(testKeyInfo, sha1, sha256, signing, encryption)
			}).ToNot(Panic())
		})
	})
})
