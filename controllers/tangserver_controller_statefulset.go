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
	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DEFAULT_STATEFULSET_PREFIX = "tangstatefulset-"
	DEFAULT_STATEFULSET_TYPE   = "StatefulSet"
	DEFAULT_PVC_STORAGE_CLASS  = ""
	DEFAULT_PVC_ACCESS_MODE    = corev1.ReadWriteOnce
	DEFAULT_PVC_STORAGE_SIZE   = "1Gi"
	DEFAULT_VOLUME_CLAIM_NAME  = "tang-keys"
)

func getStatefulSetDefaultName(cr *daemonsv1alpha1.TangServer) string {
	return DEFAULT_STATEFULSET_PREFIX + cr.Name
}

// getStatefulSet function returns correctly constructed StatefulSet
func getStatefulSet(cr *daemonsv1alpha1.TangServer) *appsv1.StatefulSet {
	labels := map[string]string{
		"app": cr.Name,
	}
	replicas := int32(cr.Spec.Replicas)
	if replicas == 0 {
		replicas = DEFAULT_REPLICA_AMOUNT
	}

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       DEFAULT_STATEFULSET_TYPE,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getStatefulSetDefaultName(cr),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: getServiceName(cr),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template:             *getStatefulSetPodTemplate(cr, labels),
			VolumeClaimTemplates: getVolumeClaimTemplates(cr),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}

// getStatefulSetPodTemplate function returns pod specification for StatefulSet
func getStatefulSetPodTemplate(cr *daemonsv1alpha1.TangServer, labels map[string]string) *corev1.PodTemplateSpec {
	lprobe := getLivenessProbe(cr)
	rprobe := getReadyProbe(cr)
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: getImageNameAndVersion(cr),
					Name:  DEFAULT_TANGSERVER_NAME,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: int32(getPodListenPort(cr)),
							Name:          DEFAULT_TANGSERVER_NAME,
						},
					},
					LivenessProbe:  lprobe,
					ReadinessProbe: rprobe,
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: getDefaultKeyPath(cr),
							Name:      DEFAULT_VOLUME_CLAIM_NAME,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: getRequests(cr),
						Limits:   getLimits(cr),
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: getSecret(cr),
				},
			},
		},
	}
}

// getVolumeClaimTemplates returns the PVC templates for StatefulSet
func getVolumeClaimTemplates(cr *daemonsv1alpha1.TangServer) []corev1.PersistentVolumeClaim {
	storageClass := getStorageClass(cr)
	storageSize := getStorageSize(cr)

	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: DEFAULT_VOLUME_CLAIM_NAME,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					DEFAULT_PVC_ACCESS_MODE,
				},
				StorageClassName: storageClass,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(storageSize),
					},
				},
			},
		},
	}
}

// getStorageClass returns the storage class for PVC
func getStorageClass(cr *daemonsv1alpha1.TangServer) *string {
	if cr.Spec.StorageClass != "" {
		return &cr.Spec.StorageClass
	}
	if DEFAULT_PVC_STORAGE_CLASS == "" {
		return nil
	}
	storageClass := DEFAULT_PVC_STORAGE_CLASS
	return &storageClass
}

// getStorageSize returns the storage size for PVC
func getStorageSize(cr *daemonsv1alpha1.TangServer) string {
	if cr.Spec.StorageSize != "" {
		return cr.Spec.StorageSize
	}
	return DEFAULT_PVC_STORAGE_SIZE
}

// getStatefulSetReadyReplicas function returns ready replicas
func getStatefulSetReadyReplicas(statefulSet *appsv1.StatefulSet) int32 {
	return statefulSet.Status.ReadyReplicas
}

// isStatefulSetReady returns a true bool if the StatefulSet has all its pods ready
func isStatefulSetReady(statefulSet *appsv1.StatefulSet) bool {
	replicas := statefulSet.Status.Replicas
	readyReplicas := statefulSet.Status.ReadyReplicas
	return replicas != 0 && replicas == readyReplicas
}

// checkStatefulSetImage returns whether the StatefulSet image is different or not
func checkStatefulSetImage(current *appsv1.StatefulSet, desired *appsv1.StatefulSet) bool {
	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}
	return false
}

// mustRedeployStatefulSet checks for cases where StatefulSet redeploy must be performed
func mustRedeployStatefulSet(new *appsv1.StatefulSet, prev *appsv1.StatefulSet) bool {
	if len(new.Spec.Template.Spec.Containers) == 0 || len(prev.Spec.Template.Spec.Containers) == 0 {
		return false
	}
	if new.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] !=
		prev.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] ||
		new.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] !=
			prev.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] ||
		new.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] !=
			prev.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] ||
		new.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] !=
			prev.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] {
		return true
	}
	return false
}
