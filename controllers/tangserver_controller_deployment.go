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
	daemonsv1alpha1 "github.com/latchset/tang-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Constants to use
const (
	DEFAULT_DEPLOYMENT_PREFIX = "tangdeployment-"
	DEFAULT_REPLICA_AMOUNT    = 1
	DEFAULT_DEPLOYMENT_TYPE   = "Deployment"
)

func getDefaultName(cr *daemonsv1alpha1.TangServer) string {
	return DEFAULT_DEPLOYMENT_PREFIX + cr.Name
}

// getDeployment function returns correctly constructed deployment
func getDeployment(cr *daemonsv1alpha1.TangServer) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	replicas := int32(cr.Spec.Replicas)
	if replicas == 0 {
		replicas = DEFAULT_REPLICA_AMOUNT
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       DEFAULT_DEPLOYMENT_TYPE,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getDefaultName(cr),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: *getPodTemplate(cr, labels),
		},
	}
}

// getDeploymentReadyReplicas function returns ready replicas
func getDeploymentReadyReplicas(deployment *appsv1.Deployment) int32 {
	return deployment.Status.ReadyReplicas
}

// isDeploymentReady returns a true bool if the deployment has all its pods ready
func isDeploymentReady(deployment *appsv1.Deployment) bool {
	replicas := deployment.Status.Replicas
	readyReplicas := deployment.Status.ReadyReplicas
	deploymentReady := false
	if replicas != 0 && replicas == readyReplicas {
		deploymentReady = true
	}
	return deploymentReady
}
