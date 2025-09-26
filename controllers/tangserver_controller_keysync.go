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
	"fmt"
	"strings"

	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// KeySyncInfo contains information needed for key synchronization
type KeySyncInfo struct {
	PrimaryPod   string
	SecondaryPod string
	Namespace    string
	DbPath       string
	TangServer   *daemonsv1alpha1.TangServer
}

// syncKeysFromPrimaryToSecondary synchronizes keys from primary pod to secondary pods
func (r *TangServerReconciler) syncKeysFromPrimaryToSecondary(keySyncInfo KeySyncInfo) error {
	GetLogInstance().Info("Syncing keys from primary to secondary",
		"Primary", keySyncInfo.PrimaryPod,
		"Secondary", keySyncInfo.SecondaryPod)

	// Copy keys from primary pod to secondary pod
	// This is a basic implementation - in production you might want more sophisticated sync
	copyCommand := fmt.Sprintf("kubectl exec %s -n %s -- tar czf - -C %s . | kubectl exec -i %s -n %s -- tar xzf - -C %s",
		keySyncInfo.PrimaryPod, keySyncInfo.Namespace, keySyncInfo.DbPath,
		keySyncInfo.SecondaryPod, keySyncInfo.Namespace, keySyncInfo.DbPath)

	GetLogInstance().Info("Key sync command", "command", copyCommand)

	// For now, we'll use a simpler approach by copying keys through the operator
	// In production, you might implement this using init containers or sidecars
	return r.copyKeysDirectly(keySyncInfo)
}

// copyKeysDirectly copies keys from primary to secondary pod using exec
func (r *TangServerReconciler) copyKeysDirectly(keySyncInfo KeySyncInfo) error {
	// Get key files from primary pod
	primaryKeys, err := r.getKeysFromPod(keySyncInfo.PrimaryPod, keySyncInfo.Namespace, keySyncInfo.DbPath)
	if err != nil {
		GetLogInstance().Error(err, "Failed to get keys from primary pod", "pod", keySyncInfo.PrimaryPod)
		return err
	}

	// Copy keys to secondary pod
	for _, keyData := range primaryKeys {
		err := r.writeKeyToPod(keySyncInfo.SecondaryPod, keySyncInfo.Namespace, keySyncInfo.DbPath, keyData)
		if err != nil {
			GetLogInstance().Error(err, "Failed to write key to secondary pod",
				"pod", keySyncInfo.SecondaryPod, "key", keyData.FileName)
			return err
		}
	}

	GetLogInstance().Info("Key synchronization completed",
		"Primary", keySyncInfo.PrimaryPod,
		"Secondary", keySyncInfo.SecondaryPod,
		"KeyCount", len(primaryKeys))

	return nil
}

// KeyData represents a Tang key file
type KeyData struct {
	FileName string
	Content  []byte
}

// getKeysFromPod retrieves all key files from a pod
func (r *TangServerReconciler) getKeysFromPod(podName, namespace, dbPath string) ([]KeyData, error) {
	GetLogInstance().Info("Getting keys from pod", "pod", podName, "path", dbPath)

	// List all .jwk files in the key directory
	listCmd := fmt.Sprintf("find %s -name '*.jwk' -type f", dbPath)
	stdout, stderr, err := podCommandExec(listCmd, DEFAULT_TANGSERVER_NAME, podName, namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Failed to list keys", "stdout", stdout, "stderr", stderr)
		return nil, err
	}

	keyFiles := strings.Split(strings.TrimSpace(stdout), "\n")
	var keyData []KeyData

	for _, keyFile := range keyFiles {
		if keyFile == "" {
			continue
		}

		// Read the key file content
		readCmd := fmt.Sprintf("cat %s", keyFile)
		content, stderr, err := podCommandExec(readCmd, DEFAULT_TANGSERVER_NAME, podName, namespace, nil)
		if err != nil {
			GetLogInstance().Error(err, "Failed to read key file", "file", keyFile, "stderr", stderr)
			continue
		}

		keyData = append(keyData, KeyData{
			FileName: keyFile,
			Content:  []byte(content),
		})
	}

	return keyData, nil
}

// writeKeyToPod writes a key file to a pod
func (r *TangServerReconciler) writeKeyToPod(podName, namespace, dbPath string, keyData KeyData) error {
	GetLogInstance().Info("Writing key to pod", "pod", podName, "file", keyData.FileName)

	// Create the key file in the secondary pod using echo instead of cat with stdin
	writeCmd := fmt.Sprintf("echo '%s' > %s", string(keyData.Content), keyData.FileName)
	_, stderr, err := podCommandExec(writeCmd, DEFAULT_TANGSERVER_NAME, podName, namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Failed to write key file", "file", keyData.FileName, "stderr", stderr)
		return err
	}

	return nil
}

// syncKeysToAllSecondaryPods synchronizes keys from primary pod to all secondary pods
func (r *TangServerReconciler) syncKeysToAllSecondaryPods(podList *corev1.PodList, keySyncInfo KeySyncInfo) error {
	if len(podList.Items) <= 1 {
		GetLogInstance().Info("No secondary pods to sync keys to")
		return nil
	}

	// Primary pod is assumed to be the first pod (index 0)
	primaryPodName := podList.Items[0].Name

	// Sync to all other pods
	for i := 1; i < len(podList.Items); i++ {
		secondaryPodName := podList.Items[i].Name

		syncInfo := KeySyncInfo{
			PrimaryPod:   primaryPodName,
			SecondaryPod: secondaryPodName,
			Namespace:    keySyncInfo.Namespace,
			DbPath:       keySyncInfo.DbPath,
			TangServer:   keySyncInfo.TangServer,
		}

		err := r.syncKeysFromPrimaryToSecondary(syncInfo)
		if err != nil {
			GetLogInstance().Error(err, "Failed to sync keys to secondary pod",
				"primary", primaryPodName, "secondary", secondaryPodName)
			// Continue with other pods even if one fails
			continue
		}
	}

	return nil
}
