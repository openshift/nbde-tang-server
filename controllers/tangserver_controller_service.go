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

	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// constants to use
const (
	DEFAULT_SERVICE_PORT   = 7500
	DEFAULT_SERVICE_TYPE   = "Service"
	DEFAULT_API_VERSION    = "v1"
	DEFAULT_SERVICE_PREFIX = "service-"
	DEFAULT_SERVICE_PROTO  = "http"
)

// getServiceName function returns service name
func getServiceName(tangserver *daemonsv1alpha1.TangServer) string {
	return DEFAULT_SERVICE_PREFIX + tangserver.Name
}

// getServicePort function returns service name
func getServicePort(tangserver *daemonsv1alpha1.TangServer) uint32 {
	servicePort := uint32(tangserver.Spec.ServiceListenPort)
	if servicePort == 0 {
		servicePort = DEFAULT_SERVICE_PORT
	}
	return servicePort
}

// getServiceType function returns the service type depending on CR information
func getServiceType(tangserver *daemonsv1alpha1.TangServer) corev1.ServiceType {
	if tangserver.Spec.ServiceType == "ClusterIP" {
		return corev1.ServiceTypeClusterIP
	}
	if tangserver.Spec.ServiceType == "NodePort" {
		return corev1.ServiceTypeNodePort
	}
	if tangserver.Spec.ServiceType == "ExternalName" {
		return corev1.ServiceTypeExternalName
	}
	return corev1.ServiceTypeLoadBalancer
}

func getClusterIP(tangserver *daemonsv1alpha1.TangServer) string {
	return tangserver.Spec.ClusterIP
}

// getService function returns correctly created service
func getService(tangserver *daemonsv1alpha1.TangServer) *corev1.Service {
	GetLogInstance().Info("getService")
	labels := map[string]string{
		"app": tangserver.Name,
	}
	servicePort := getServicePort(tangserver)
	GetLogInstance().Info("Listening Port", "servicePort", servicePort)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: DEFAULT_API_VERSION,
			Kind:       DEFAULT_SERVICE_TYPE,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getServiceName(tangserver),
			Namespace: tangserver.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     getServiceType(tangserver),
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       DEFAULT_SERVICE_PROTO,
					Port:       int32(servicePort),
					TargetPort: intstr.FromInt(int(getPodListenPort(tangserver))),
				},
			},
			ClusterIP: getClusterIP(tangserver),
		},
	}
}

// getService function returns correctly created service
func getServiceURL(tangserver *daemonsv1alpha1.TangServer) string {
	return DEFAULT_SERVICE_PROTO + "://" + getServiceName(tangserver) + "." + tangserver.Namespace + ":" + fmt.Sprint(getServicePort(tangserver)) + "/adv"
}

// getServiceIpURL function returns correctly created service
func getServiceIpURL(tangserver *daemonsv1alpha1.TangServer, ip string) string {
	return DEFAULT_SERVICE_PROTO + "://" + ip + ":" + fmt.Sprint(getServicePort(tangserver)) + "/adv"
}

// getExternalServiceURL function returns correctly created service
func getExternalServiceURL(tangserver *daemonsv1alpha1.TangServer, balancer corev1.LoadBalancerIngress) string {
	if len(balancer.Hostname) > 0 {
		return DEFAULT_SERVICE_PROTO + "://" + balancer.Hostname + ":" + fmt.Sprint(getServicePort(tangserver)) + "/adv"
	} else {
		return DEFAULT_SERVICE_PROTO + "://" + balancer.IP + ":" + fmt.Sprint(getServicePort(tangserver)) + "/adv"
	}
}
