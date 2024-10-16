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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DEFAULT_POD_RUNNING_PORT = 8080
const DEFAULT_TANGSERVER_NAME = "tangserver"
const DEFAULT_TANGSERVER_PVC_NAME = "tangserver-pvc"
const DEFAULT_TANGSERVER_SECRET = "tangserversecret"

// getPodListenPort function returns the internal port where tangserver will listen
func getPodListenPort(cr *daemonsv1alpha1.TangServer) int32 {
	if cr.Spec.PodListenPort != 0 {
		return cr.Spec.PodListenPort
	}
	return DEFAULT_POD_RUNNING_PORT
}

// getSecret function returns the internal port where tangserver will listen
func getSecret(cr *daemonsv1alpha1.TangServer) string {
	if cr.Spec.Secret != "" {
		return cr.Spec.Secret
	}
	return DEFAULT_TANGSERVER_SECRET
}

// getPersistentVolumeClaim function returns the internal port where tangserver will listen
func getPersistentVolumeClaim(cr *daemonsv1alpha1.TangServer) string {
	if cr.Spec.PersistentVolumeClaim != "" {
		return cr.Spec.PersistentVolumeClaim
	}
	return DEFAULT_TANGSERVER_PVC_NAME
}

// getRequests returns the information based on the information in the CRD
func getRequests(cr *daemonsv1alpha1.TangServer) corev1.ResourceList {
	requests := make(map[corev1.ResourceName]resource.Quantity)
	if cr.Spec.ResourcesRequest.Cpu != "" {
		cpu, e := resource.ParseQuantity(cr.Spec.ResourcesRequest.Cpu)
		if e == nil {
			requests[corev1.ResourceCPU] = cpu
		}
	}
	if cr.Spec.ResourcesRequest.Memory != "" {
		mem, e := resource.ParseQuantity(cr.Spec.ResourcesRequest.Memory)
		if e == nil {
			requests[corev1.ResourceMemory] = mem
		}
	}
	return requests
}

// getLimits returns the information based on the information in the CRD
func getLimits(cr *daemonsv1alpha1.TangServer) corev1.ResourceList {
	limits := make(map[corev1.ResourceName]resource.Quantity)
	if cr.Spec.ResourcesLimit.Cpu != "" {
		cpu, e := resource.ParseQuantity(cr.Spec.ResourcesLimit.Cpu)
		if e == nil {
			limits[corev1.ResourceCPU] = cpu
		}
	}
	if cr.Spec.ResourcesLimit.Memory != "" {
		mem, e := resource.ParseQuantity(cr.Spec.ResourcesLimit.Memory)
		if e == nil {
			limits[corev1.ResourceMemory] = mem
		}
	}
	return limits
}

// getPodTemplate function returns pod specification according to tangserver spec
func getPodTemplate(cr *daemonsv1alpha1.TangServer, labels map[string]string) *corev1.PodTemplateSpec {
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
							Name:      getPersistentVolumeClaim(cr),
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: getRequests(cr),
						Limits:   getLimits(cr),
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: getPersistentVolumeClaim(cr),
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: getPersistentVolumeClaim(cr),
						},
					},
				},
			},
			// TODO: Check how to change Restart Policy
			RestartPolicy: corev1.RestartPolicyAlways,
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: getSecret(cr),
				},
			},
		},
	}
}
