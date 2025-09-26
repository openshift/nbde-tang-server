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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TangServerSpec defines the desired state of TangServer
type TangServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// KeyPath is field of TangServer. It allows to specify the path where keys will be generated
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Key Path"
	// +optional
	KeyPath string `json:"keyPath,omitempty"`

	// Replicas is the Tang Server amount to bring up
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Amount of replicas to launch"
	Replicas int32 `json:"replicas"`

	// Persistent Volume Claim to store the keys
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Persistent Volume Claim to attach to (default:tangserver-pvc)"
	// +optional
	PersistentVolumeClaim string `json:"persistentVolumeClaim,omitempty"`

	// Image is the base container image of the TangServer to use
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Image of Container to deploy"
	// +optional
	Image string `json:"image,omitempty"`

	// Version is the version of the TangServer container to use (empty=>latest)
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Image Version of Container to deploy"
	// +optional
	Version string `json:"version,omitempty"`

	// HealthScript is the script to run for healthiness/readiness
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Health Script to execute"
	// +optional
	HealthScript string `json:"healthScript,omitempty"`

	// PodListenPort is the port where pods will listen for traffic
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Port where Pod will listen "
	// +optional
	PodListenPort int32 `json:"podListenPort,omitempty"`

	// Secret is the secret name to use to download image appropriately
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret name to use for container download"
	// +optional
	Secret string `json:"secret,omitempty"`

	// ServiceListenPort is the port where service will listen for traffic
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Port where service will listen"
	// +optional
	ServiceListenPort int32 `json:"serviceListenPort,omitempty"`

	// ResourceRequest is the resource request to perform for each pod
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resources Request for Tang Server"
	// +optional
	ResourcesRequest ResourcesRequest `json:"resourcesRequest,omitempty"`

	// ResourceLimit is the resource limit to perform for each pod
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resources Limit for Tang Server"
	// +optional
	ResourcesLimit ResourcesLimit `json:"resourcesLimit,omitempty"`

	// KeyRefreshInterval
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Refresh Interval to update key status"
	// +optional
	KeyRefreshInterval uint32 `json:"keyRefreshInterval,omitempty"`

	// HiddenKeys
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Hidden Keys contains a list with the keys (with sha1 or sha256) to hide"
	// +optional
	HiddenKeys []TangServerHiddenKeys `json:"hiddenKeys,omitempty"`

	// RequiredActiveKeyPairs
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Required Active Key Pairs (1 by default)"
	// +optional
	RequiredActiveKeyPairs uint32 `json:"requiredActiveKeyPairs,omitempty"`

	// ServiceType
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ServiceType (LoadBalancer by default)"
	// +optional
	ServiceType string `json:"serviceType,omitempty"`

	// ClusterIP
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ClusterIP (empty by default)"
	// +optional
	ClusterIP string `json:"clusterIP,omitempty"`

	// StorageClass
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="StorageClass for PVC (empty for default)"
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

	// StorageSize
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage size for PVC (1Gi by default)"
	// +optional
	StorageSize string `json:"storageSize,omitempty"`
}

// ResourcesRequest contains the struct to provide resources requests to Tang Server
type ResourcesRequest struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// ResourcesLimit contains the struct to provide resources limit to Tang Server
type ResourcesLimit struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// TangServerActiveKeys defines the active keys in a Tang Server
type TangServerActiveKeys struct {
	// Active Key sha1
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Active Key SHA1"
	// +optional
	Sha1 string `json:"sha1,omitempty"`
	// Active Key sha256
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Active Key SHA256"
	// +optional
	Sha256 string `json:"sha256,omitempty"`
	// Active Key Generation Time
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Active Key Generation Time"
	Generated string `json:"generated,omitempty"`
	// FileName provides information about the file name corresponding to the key
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Active Key file name"
	// +optional
	FileName string `json:"fileName,omitempty"`
}

// TangServerHiddenKeys defines the hidden keys in a Tang Server
type TangServerHiddenKeys struct {
	// Hidden Key sha1
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Hidden Key SHA1"
	// +optional
	Sha1 string `json:"sha1,omitempty"`
	// Hidden Key sha256
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Hidden Key SHA256"
	// +optional
	Sha256 string `json:"sha256,omitempty"`
	// Hidden Key Hiding Time
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Hidden Key Generation Time"
	Generated string `json:"generated,omitempty"`
	// Hidden Key Generation Time
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Hidden Key Hidden Time"
	Hidden string `json:"hidden,omitempty"`
	// FileName provides information about the file name corresponding to the key
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Hidden Key file name"
	// +optional
	FileName string `json:"fileName,omitempty"`
}

// TangServerStatus defines the observed state of TangServer
type TangServerStatus struct {
	// TangServerError collects error on Tang Operator creation
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Tang Server Error"
	// +optional
	TangServerError TangServerStatusError `json:"tangServerError,omitempty"`
	// ActiveKeys provides information about the Active Keys in the Tang Server CR
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Tang Server Active Keys"
	// +optional
	ActiveKeys []TangServerActiveKeys `json:"activeKeys,omitempty"`
	// HiddenKeys provides information about the Hidden Keys in the Tang Server CR
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Tang Server Hidden Keys"
	// +optional
	HiddenKeys []TangServerHiddenKeys `json:"hiddenKeys,omitempty"`
	// Tang Server Running provides information about the Running Replicas
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Tang Server Running Replicas"
	// +optional
	Running int32 `json:"running"`
	// Tang Server Ready provides information about the Ready Replicas
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Tang Server Ready Replicas"
	// +optional
	Ready int32 `json:"ready"`
	// Tang Server Service External URL provides information about the External Service URL
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text",displayName="Tang Server External URL"
	// +optional
	ServiceExternalURL string `json:"serviceExternalURL,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="KeyPath",type="string",JSONPath=".spec.keypath",description="Directory to use for key generation"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Replicas to launch for a particular deployment"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.replicas",description="Container Image to use"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="Version of the Container Image to use"
// +kubebuilder:printcolumn:name="HealthScript",type="string",JSONPath=".spec.healthscript",description="Health Script to execute"
// +kubebuilder:printcolumn:name="PodListenPort",type="integer",JSONPath=".spec.podlistenport",description="Port where each Pod will listen"
// +kubebuilder:printcolumn:name="Secret",type="string",JSONPath=".spec.secret",description="Secret name to use in case it is necessary"
// +kubebuilder:printcolumn:name="ServiceListenPort",type="integer",JSONPath=".spec.podlistenport",description="Port where each Service will listen"
// TangServer is the Schema for the tangservers API
type TangServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TangServerSpec   `json:"spec,omitempty"`
	Status TangServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TangServerList contains a list of TangServer
type TangServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TangServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TangServer{}, &TangServerList{})
}
