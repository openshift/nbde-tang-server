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

	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
)

const KEY_STATUS_FILE_NAME = "key_status.txt"

type KeyAssociation struct {
	Sha1          string `json:"-"`
	Sha256        string `json:"-"`
	SigningKey    string `json:"signing"`
	EncriptionKey string `json:"encryption"`
}

type KeyAssociationSha1Map map[string]KeyAssociation
type KeyAssociationSha256Map map[string]KeyAssociation
type KeyAssociationShaMap map[string]KeyAssociation

type KeyAssociationMap struct {
	KeyStatusSha1Map   KeyAssociationSha1Map   `json:"sha1"`
	KeyStatusSha256Map KeyAssociationSha256Map `json:"sha256"`
}

type KeySelectiveMap map[string]string

func keyStatusFile() string {
	return KEY_STATUS_FILE_NAME
}

func keyStatusFilePathWithTangServer(ts *daemonsv1alpha1.TangServer) string {
	return getDefaultKeyPath(ts) + "/" + KEY_STATUS_FILE_NAME
}

func keyStatusFilePath(k KeyAssociationInfo) string {
	return getDefaultKeyPath(k.KeyInfo.TangServer) + "/" + KEY_STATUS_FILE_NAME
}

func keyStatusLockFilePath(k KeyAssociationInfo) string {
	return getDefaultKeyPath(k.KeyInfo.TangServer) + "/" + KEY_STATUS_FILE_NAME + ".lock"
}

func deleteHiddenKeysSelectively(keepKeys KeySelectiveMap, keyinfo KeyObtainInfo) error {
	// If Key Status File Exist, unmarshal it
	statusFile := keyStatusFilePathWithTangServer(keyinfo.TangServer)
	command := "cat " + statusFile
	GetLogInstance().Info("deleteHiddenKeysSelectively", "Keys to keep", keepKeys)
	stdo, _, e := podCommandExec(command, "", keyinfo.PodName, keyinfo.Namespace, nil)
	if e != nil {
		GetLogInstance().Error(e, "deleteHiddenKeysSelectively: Unable to read status file", "statusFile", statusFile)
	} else {
		var KeyStatusMap KeyAssociationMap
		if err := json.Unmarshal([]byte(stdo), &KeyStatusMap); err != nil {
			GetLogInstance().Error(err, "deleteHiddenKeysSelectively: Unable to unmarshal status file", "Status File", statusFile, "JSON content", stdo)
			return err
		}
		// Join Both Sha Maps
		joinedMaps := make(KeyAssociationShaMap, 0)
		for k1, v1 := range KeyStatusMap.KeyStatusSha1Map {
			joinedMaps[k1] = v1
		}
		for k2, v2 := range KeyStatusMap.KeyStatusSha256Map {
			joinedMaps[k2] = v2
		}

		for k, v := range joinedMaps {
			if _, found := keepKeys[k]; !found {
				//remove only if it is a hidden key
				for _, hsk := range keyinfo.TangServer.Status.HiddenKeys {
					if k == hsk.Sha1 || k == hsk.Sha256 {
						//delete signing and encryption hidden keys!!
						GetLogInstance().Info("deleteHiddenKeysSelectively: deletePodFiles", "Key Association", v, "SHA1/SHA256 not found", k)
						if e := deletePodFile(keyinfo, v); e != nil {
							GetLogInstance().Error(e, "deleteHiddenKeysSelectively: Error Deleting Key Association Files", "Key Association", v)
						} else {
							GetLogInstance().Info("deleteHiddenKeysSelectively: Keys Deleted Correctly", "Key Association", v)
						}
					}
				}
			}
		}
	}
	return nil
}

func dumpKeyAssociation(k KeyAssociationInfo) error {
	updateForbiddenMap(keyStatusFile())
	// If lock file exists, do nothing
	keyStatusLockFilePath := keyStatusLockFilePath(k)
	command := "test -f " + keyStatusLockFilePath
	_, _, err := podCommandExec(command, "", k.KeyInfo.PodName, k.KeyInfo.Namespace, nil)
	if err == nil {
		GetLogInstance().Info("Lock operation in progress")
		return nil
	}
	// Lock
	command = "touch " + keyStatusLockFilePath
	_, _, err = podCommandExec(command, "", k.KeyInfo.PodName, k.KeyInfo.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to lock status file")
		return err
	}
	var KeyStatusMap KeyAssociationMap
	if KeyStatusMap.KeyStatusSha1Map == nil {
		KeyStatusMap.KeyStatusSha1Map = make(KeyAssociationSha1Map, 1)
	}
	if KeyStatusMap.KeyStatusSha256Map == nil {
		KeyStatusMap.KeyStatusSha256Map = make(KeyAssociationSha256Map, 1)
	}
	statusFile := keyStatusFilePath(k)
	// If Key Status File Exist, unmarshal it
	command = "cat " + statusFile
	stdo, _, e := podCommandExec(command, "", k.KeyInfo.PodName, k.KeyInfo.Namespace, nil)
	if e == nil {
		GetLogInstance().Info("Updating status map with key status file")
		if err = json.Unmarshal([]byte(stdo), &KeyStatusMap); err != nil {
			GetLogInstance().Error(err, "Unable to unmarshal status file", "Status File", statusFile, "JSON content", stdo)
		}
	}
	delete(KeyStatusMap.KeyStatusSha1Map, k.KeyAssoc.Sha1)
	delete(KeyStatusMap.KeyStatusSha256Map, k.KeyAssoc.Sha256)
	KeyStatusMap.KeyStatusSha1Map[k.KeyAssoc.Sha1] = k.KeyAssoc
	KeyStatusMap.KeyStatusSha256Map[k.KeyAssoc.Sha256] = k.KeyAssoc
	keyStatus, err := json.Marshal(KeyStatusMap)
	if err != nil {
		GetLogInstance().Error(err, "Error on KeyStatusMap marshalling", "file", statusFile, "keyStatusMap", KeyStatusMap)
	}
	GetLogInstance().Info("Dumping key status to file", "file", statusFile, "keyStatus", string(keyStatus))
	err = dumpKeyStatusFileWithEchoRedirection(statusFile, keyStatus, k.KeyInfo.PodName, k.KeyInfo.Namespace)
	if err != nil {
		GetLogInstance().Error(err, "Error Dumping Key Status File", "file", statusFile, "keyStatus", string(keyStatus))
	}

	// Unlock
	command = "rm -fr " + keyStatusLockFilePath
	_, _, err = podCommandExec(command, "", k.KeyInfo.PodName, k.KeyInfo.Namespace, nil)
	if err != nil {
		GetLogInstance().Error(err, "Unable to delete lock status file")
		return err
	}
	return err
}
