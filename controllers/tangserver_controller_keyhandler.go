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
	"strings"

	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
)

type SHAType uint8
type FileModType uint8
type KeyAdvertisingType uint8

const (
	UNKNOWN_SHA SHAType = iota
	SHA256
	SHA1
)

const (
	UNKNOWN_MOD FileModType = iota
	CREATION
	MODIFICATION
)

const (
	UNKNOWN_ADVERTISED KeyAdvertisingType = iota
	ALL_KEYS
	ONLY_ADVERTISED
	ONLY_UNADVERTISED
)

const DEFAULT_DEPLOYMENT_KEY_PATH = "/var/db/tang"

var FORBIDDEN_PATH_MAP = map[string]string{
	".":          "FORBIDDEN",
	"..":         "FORBIDDEN",
	"lost+found": "FORBIDDEN",
}

type KeyObtainInfo struct {
	PodName    string
	Namespace  string
	DbPath     string
	TangServer *daemonsv1alpha1.TangServer
}

type KeyAssociationInfo struct {
	KeyInfo  *KeyObtainInfo
	KeyAssoc KeyAssociation
}

type KeyRotateInfo struct {
	KeyInfo     *KeyObtainInfo
	KeyFileName string
}

// getDefaultKeyPath returns directory where keys are dumped, typically /var/db/tang
func getDefaultKeyPath(cr *daemonsv1alpha1.TangServer) string {
	if cr.Spec.KeyPath != "" {
		return cr.Spec.KeyPath
	}
	return DEFAULT_DEPLOYMENT_KEY_PATH
}

// keyToAdvertise returns if a key is to be advertised (is a signing key)
func keyToAdvertise(keyInfo KeyObtainInfo, path string) bool {
	command := "jose jwk use --input " + path + " --required --use=verify"
	_, _, notAdvertisable := podCommandExec(command, "", keyInfo.PodName, keyInfo.Namespace, nil)
	if notAdvertisable != nil {
		GetLogInstance().Info("Key not advertisable", "key path", path)
		return false
	}
	GetLogInstance().Info("Key advertisable", "key path", path)
	return true
}

// ignoreKey function checks if key must be ignored
func ignoreKey(keyInfo KeyObtainInfo, advertised KeyAdvertisingType, keypath string) bool {
	if keyToAdvertise(keyInfo, keypath) {
		if advertised == ONLY_UNADVERTISED {
			GetLogInstance().Info("Key ignored", "key path", keypath)
			return true
		}
	} else {
		if advertised == ONLY_ADVERTISED {
			GetLogInstance().Info("Key ignored", "key path", keypath)
			return true
		}
	}
	GetLogInstance().Info("Key not ignored", "key path", keypath)
	return false
}

// writeStatusFile function
func writeStatusFile(keyInfo KeyObtainInfo, sha1 string, sha256 string, statusSigning string, statusEncryption string) error {
	if len(sha1) > 0 && len(sha256) > 0 && len(statusSigning) > 0 && len(statusEncryption) > 0 {
		GetLogInstance().Info("writeStatusFile", "sha1", sha1, "sha256", sha256, "statusSigning", statusSigning, "statusEncryption", statusEncryption)
		k := KeyAssociationInfo{
			KeyInfo: &keyInfo,
			KeyAssoc: KeyAssociation{
				Sha1:          sha1,
				Sha256:        sha256,
				SigningKey:    statusSigning,
				EncriptionKey: statusEncryption,
			},
		}
		GetLogInstance().Info("Dumping Key Association", "Key Association", k)
		return dumpKeyAssociation(k)
	}
	return nil
}

func updateForbiddenMap(forbiddenKey string) {
	FORBIDDEN_PATH_MAP[forbiddenKey] = "FORBIDDEN"
}

// readActiveKeys function return active key list
func readActiveKeys(keyInfo KeyObtainInfo, onlyAdvertised KeyAdvertisingType) ([]daemonsv1alpha1.TangServerActiveKeys, error) {
	command := "ls " + keyInfo.DbPath
	stdo, stde, err := podCommandExec(command, "", keyInfo.PodName, keyInfo.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "stderror", stde, "podname", keyInfo.PodName, "namespace", keyInfo.Namespace)
	} else {
		GetLogInstance().Info("Executed active keys retrieval command correctly", "Active keys:", stdo)
		keys := strings.Split(stdo, "\n")
		activeKeys := make([]daemonsv1alpha1.TangServerActiveKeys, 0)
		var statusSigning string
		var statusEncryption string
		var ignoredKey bool
		var sha1 string
		var sha256 string
		for _, k := range keys {
			if len(k) > 0 {
				if _, forbidden := FORBIDDEN_PATH_MAP[k]; forbidden {
					continue
				}
				k = strings.TrimLeft(strings.TrimRight(k, "\n"), "\n")
				k = strings.TrimLeft(strings.TrimRight(k, "\r"), "\r")
				fpath := keyInfo.DbPath + "/" + k
				if ignoreKey(keyInfo, onlyAdvertised, fpath) {
					statusEncryption = fpath
					ignoredKey = true
				} else {
					statusSigning = fpath
					ignoredKey = false
				}
				if !ignoredKey {
					sha1 = getSHA(SHA1, keyInfo, fpath)
					sha256 = getSHA(SHA256, keyInfo, fpath)
					activeKeys = append(activeKeys, daemonsv1alpha1.TangServerActiveKeys{
						Sha1:      sha1,
						Sha256:    sha256,
						Generated: getLastTime(CREATION, keyInfo, fpath),
						FileName:  k,
					})
				}
				if err := writeStatusFile(keyInfo, sha1, sha256, statusSigning, statusEncryption); err != nil {
					GetLogInstance().Error(err, "Unable to write status file", "keyInfo", keyInfo)
				}
			}
		}
		return activeKeys, nil
	}
	return nil, err
}

// readHiddenKeys function return hidden key list
func readHiddenKeys(keyInfo KeyObtainInfo, onlyAdvertised KeyAdvertisingType) ([]daemonsv1alpha1.TangServerHiddenKeys, error) {
	command := "ls -a " + keyInfo.DbPath + "/"
	stdo, stde, err := podCommandExec(command, "", keyInfo.PodName, keyInfo.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "stdo", stdo, "stderror", stde, "podname", keyInfo.PodName, "namespace", keyInfo.Namespace)
	} else {
		GetLogInstance().Info("Executed hidden keys retrieval command correctly", "Hidden keys:", stdo)
		keys := strings.Split(stdo, "\n")
		hiddenKeys := make([]daemonsv1alpha1.TangServerHiddenKeys, 0)
		var statusSigning string
		var statusEncryption string
		var ignoredKey bool
		var sha1 string
		var sha256 string
		for _, k := range keys {
			if len(k) > 0 {
				if _, forbidden := FORBIDDEN_PATH_MAP[k]; forbidden {
					continue
				}
				if k[0] == '.' {
					k = strings.TrimLeft(strings.TrimRight(k, "\n"), "\n")
					k = strings.TrimLeft(strings.TrimRight(k, "\r"), "\r")
					fpath := keyInfo.DbPath + "/" + k
					if ignoreKey(keyInfo, onlyAdvertised, fpath) {
						statusEncryption = fpath
						ignoredKey = true
					} else {
						statusSigning = fpath
						ignoredKey = false
					}
					if !ignoredKey {
						sha1 = getSHA(SHA1, keyInfo, fpath)
						sha256 = getSHA(SHA256, keyInfo, fpath)
						hiddenKeys = append(hiddenKeys, daemonsv1alpha1.TangServerHiddenKeys{
							Sha1:      sha1,
							Sha256:    sha256,
							Generated: getCreationTimeFromKeys(keyInfo, sha1),
							Hidden:    getLastTime(MODIFICATION, keyInfo, fpath),
							FileName:  k,
						})
					}
					if err := writeStatusFile(keyInfo, sha1, sha256, statusSigning, statusEncryption); err != nil {
						GetLogInstance().Error(err, "Unable to write status file", "keyInfo", keyInfo)
					}
				}
			}
		}
		return hiddenKeys, nil
	}
	return nil, err
}

// getCreationTimeFromKeys function returns creation time for an active or hidden key with its sha1
func getCreationTimeFromKeys(keyInfo KeyObtainInfo, sha1 string) string {
	for _, k := range keyInfo.TangServer.Status.ActiveKeys {
		if k.Sha1 == sha1 {
			return k.Generated
		}
	}
	// Check if its already stored
	return getCreationTimeFromHiddenKey(keyInfo, sha1)
}

// getCreationTimeFromHiddenKey function returns creation time for an active key with its sha1
func getCreationTimeFromHiddenKey(keyInfo KeyObtainInfo, sha1 string) string {
	for _, k := range keyInfo.TangServer.Status.HiddenKeys {
		if k.Sha1 == sha1 {
			return k.Generated
		}
	}
	GetLogInstance().Info("Unable to obtain creation time")
	return "UNKNOWN_CREATION_TIME"
}

// createNewPairOfKeys function creates new pair of keys (via /usr/libexec/tangd-keygen)
func createNewPairOfKeys(k KeyObtainInfo) error {
	command := "/usr/libexec/tangd-keygen " + k.DbPath + "/"
	stdo, stde, err := podCommandExec(command, "", k.PodName, k.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "stdo", stdo, "stderror", stde, "podname", k.PodName, "namespace", k.Namespace)
	}
	return err
}

// rotateUnadvertisedKeys function rotate key file, moving it to hidden file
// TODO: Rotate the key corresponding to a particular signing key
//
//	Right now, all unadvertised keys will be rotated
func rotateUnadvertisedKeys(krinfo KeyRotateInfo) error {
	var ge error
	GetLogInstance().Info("rotateUnadvertisedKeys", "Advertised Key Info", krinfo.KeyFileName)
	keys, e := readActiveKeys(*krinfo.KeyInfo, ONLY_UNADVERTISED)
	if e != nil {
		GetLogInstance().Error(e, "Unable to read unadvertised keys", "Key Rotate Info", krinfo, "podname", krinfo.KeyInfo.PodName, "namespace", krinfo.KeyInfo.Namespace)
		return e
	}
	ge = nil
	for _, uk := range keys {
		rk := KeyRotateInfo{
			KeyInfo:     krinfo.KeyInfo,
			KeyFileName: uk.FileName,
		}
		e := rotateKey(rk)
		if ge == nil && e != nil {
			GetLogInstance().Error(e, "Error rotating unadvertised key", "Rotate Key Info", rk)
			ge = e
		}
	}
	return ge
}

// dumpKeyStatusFileWithEchoRedirection receives the key in string format and the file where it is to be dumped, and dumps it
func dumpKeyStatusFileWithEchoRedirection(keyFile string, fileContent []byte, podName string, namespace string) error {
	command := `echo '` + string(fileContent) + `' > ` + keyFile
	stdo, stde, err := podCommandExec(command, "", podName, namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "podname", podName, "namespace", namespace, "stdo", stdo, "stde", stde)
	} else {
		GetLogInstance().Info("Command executed successfully", "command", command, "podname", podName, "namespace", namespace, "stdo", stdo, "stde", stde)
	}
	return err
}

// rotateKey function rotate key file, moving it to hidden file
func rotateKey(k KeyRotateInfo) error {
	command := "mv " + k.KeyInfo.DbPath + "/" + k.KeyFileName + " " + k.KeyInfo.DbPath + "/." + k.KeyFileName
	stdo, stde, err := podCommandExec(command, "", k.KeyInfo.PodName, k.KeyInfo.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "stdo", stdo, "stderror", stde, "podname", k.KeyInfo.PodName, "namespace", k.KeyInfo.Namespace)
	} else {
		GetLogInstance().Info("Move file command executed correctly", "command", command, "podname", k.KeyInfo.PodName, "namespace", k.KeyInfo.Namespace)
	}
	return err
}

// getSHA function returns SHA1 or SHA256 of the file provided in the parameters
func getSHA(shaType SHAType, keyInfo KeyObtainInfo, filePath string) string {
	alg := "Unknown"
	switch shaType {
	case SHA1:
		alg = "S1"
	case SHA256:
		alg = "S256"
	}
	command := "jose jwk thp -a" + alg + " -i " + filePath
	stdo, stde, err := podCommandExec(command, "", keyInfo.PodName, keyInfo.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "stdo", stdo, "stderror", stde, "podname", keyInfo.PodName, "namespace", keyInfo.Namespace)
		return ""
	}
	return stdo
}

// getLastTime indicates last creation/modficiation time of the file
func getLastTime(fmod FileModType, keyInfo KeyObtainInfo, filePath string) string {
	command := "stat -c "
	ftype := ""
	switch fmod {
	case CREATION:
		ftype += "'%y'"
	case MODIFICATION:
		ftype += "'%z'"
	}
	command += ftype + " " + filePath
	stdo, stde, err := podCommandExec(command, "", keyInfo.PodName, keyInfo.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "stdo", stdo, "stderror", stde, "podname", keyInfo.PodName, "namespace", keyInfo.Namespace)
		return ""
	}
	//return strings.TrimLeft(strings.TrimRight(stdo, "\n"), "\n")
	return strings.TrimRight(strings.TrimLeft(strings.TrimLeft(strings.TrimRight(stdo, "\n"), "\n"), "'"), "'")
}

// deleteAllHiddenKeys function return active key list
func deleteAllHiddenKeys(keyInfo KeyObtainInfo) bool {
	if len(keyInfo.TangServer.Status.ActiveKeys) > 0 {
		command := "rm -frv"
		ahk, e := readHiddenKeys(keyInfo, ALL_KEYS)
		if e != nil {
			GetLogInstance().Error(e, "Unable to read hidden keys", "podname", keyInfo.PodName, "namespace", keyInfo.Namespace)
			return false
		}
		for _, kf := range ahk {
			command += " " + keyInfo.DbPath + "/" + kf.FileName
		}
		stdo, stde, err := podCommandExec(command, "", keyInfo.PodName, keyInfo.Namespace, nil)
		GetLogInstance().Info("Executing command in Pod", "command", command, "podname", keyInfo.PodName)
		if err != nil {
			GetLogInstance().Error(err, "Unable to execute command in Pod", "command", command, "stdo", stdo, "stderror", stde, "podname", keyInfo.PodName, "namespace", keyInfo.Namespace)
			return false
		} else {
			GetLogInstance().Info("Command correctly executed", "output", stdo, "error", stde)
		}
	}
	return true
}
