#!/bin/bash
#
#   Copyright [2023] [sarroutb (at) redhat.com]
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#
counter=0
oc apply -f operator_configs/minimal-keyretrieve
while true;
do
  echo "======================  $counter =================="
  oc -n nbde describe tangservers.nbde.openshift.io
  ./tools/api_tools/key_rotate.sh -n nbde
  sleep 10
  oc -n nbde describe tangservers.nbde.openshift.io
  oc apply -f operator_configs/minimal-keyretrieve-deletehiddenkeys
  sleep 5
  oc -n nbde describe tangservers.nbde.openshift.io
  echo "====================== /$counter =================="
  sleep 5
  ((counter+=1))
done
echo
