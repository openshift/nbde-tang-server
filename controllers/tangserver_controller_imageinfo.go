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
)

const DEFAULT_APP_IMAGE = "registry.redhat.io/rhel9/tang"
const DEFAULT_APP_VERSION = "latest"

func getCompleteImageNameAndVersion(appImage string, appVersion string) string {
	return appImage + ":" + appVersion
}

// getImageNameAndVersionName will return the image to use, or the default one
// if no one is specified in the CRD
func getImageNameAndVersion(cr *daemonsv1alpha1.TangServer) string {
	appImage := DEFAULT_APP_IMAGE
	appVersion := DEFAULT_APP_VERSION
	if cr.Spec.Image != "" {
		appImage = cr.Spec.Image
	}
	if cr.Spec.Version != "" {
		appVersion = cr.Spec.Version
	}
	return getCompleteImageNameAndVersion(appImage, appVersion)
}
