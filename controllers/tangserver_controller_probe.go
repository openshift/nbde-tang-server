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
)

const DEFAULT_DEPLOYMENT_HEALTH_CHECK = "/usr/bin/tangd-health-check"

const DEFAULT_READY_INITIALDELAYSECONDS = 5
const DEFAULT_READY_TIMEOUT_SECONDS = 5
const DEFAULT_READY_PERIOD_SECONDS = 15

const DEFAULT_LIVENESS_INITIALDELAYSECONDS = 27
const DEFAULT_LIVENESS_TIMEOUT_SECONDS = 5
const DEFAULT_LIVENESS_PERIOD_SECONDS = 16

// getReadyProbe function returns appropriate probe taking into account tangserver spec
func getReadyProbe(cr *daemonsv1alpha1.TangServer) *corev1.Probe {
	healthScript := DEFAULT_DEPLOYMENT_HEALTH_CHECK
	if cr.Spec.HealthScript != "" {
		healthScript = cr.Spec.HealthScript
	}
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					healthScript,
				},
			},
		},
		InitialDelaySeconds: DEFAULT_READY_INITIALDELAYSECONDS,
		TimeoutSeconds:      DEFAULT_READY_TIMEOUT_SECONDS,
		PeriodSeconds:       DEFAULT_READY_PERIOD_SECONDS,
	}
}

// getLivenessProbe function returns appropriate probe taking into account tangserver spec
func getLivenessProbe(cr *daemonsv1alpha1.TangServer) *corev1.Probe {
	healthScript := DEFAULT_DEPLOYMENT_HEALTH_CHECK
	if cr.Spec.HealthScript != "" {
		healthScript = cr.Spec.HealthScript
	}
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					healthScript,
				},
			},
		},
		InitialDelaySeconds: DEFAULT_LIVENESS_INITIALDELAYSECONDS,
		TimeoutSeconds:      DEFAULT_LIVENESS_TIMEOUT_SECONDS,
		PeriodSeconds:       DEFAULT_LIVENESS_PERIOD_SECONDS,
	}
}
