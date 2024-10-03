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
namespace=""
wait_secs=15
wait_up_secs=30

usage() {
  echo
  echo "Usage:"
  echo
  echo "$1 -n namespace [-c k8s_client] [-v (verbose)]"
  echo
  exit "$2"
}

while getopts "d:n:c:p:i:u:hv" arg
do
case "${arg}" in
  d) device=${OPTARG}
  ;;
  n) namespace=${OPTARG}
  ;;
  c) oc_client=${OPTARG}
  ;;
  p) project_root=${OPTARG}
  ;;
  i) clevis_ip=${OPTARG}
  ;;
  u) wait_up_secs=${OPTARG}
  ;;
  h) usage "$0" 0
  ;;
  v) set -x
  ;;
  *) usage "$0" 1
  ;;
esac
done

this_path=$(dirname "${0}")
test -z "${namespace}" && namespace="nbde"
test -z "${oc_client}" && oc_client="oc"
test -z "${project_root}" && project_root="${this_path}/../.."
test -z "${clevis_ip}" && clevis_ip="192.168.122.126"
test -z "${device}" && device="/dev/vda2"

prompt_command() {
    echo "-------------------------------------------------"
    echo "$1"
    echo -n "Execute command [$1]? [Y/n]:"
    read -r -n 1 input
    if [ "${input}" == "N" ] || [ "${input}" == "n" ]; then
	return
    fi
    ${1}
}

prompt_command_vm_manually() {
    echo "-------------------------------------------------"
    echo "$1"
    echo -n "Please, execute command [$1] in VM:[$2] manually, press any key to continue ..."
    read -r -n 1 input
}

# Step 1) Install operator
prompt_command "operator-sdk run bundle --timeout 5m quay.io/sec-eng-special/nbde-tang-server-bundle:latest"

# Step 2) Install key management configuration
prompt_command "${oc_client} apply -f ${project_root}/operator_configs/minimal-keyretrieve"

counter=0
while [ $counter -ne "${wait_up_secs}" ];
do
  ((counter++))
  echo -n -e "\rSleeping until pod and service is up and running ... ${counter}/${wait_up_secs}"
  sleep 1
done
echo

# Step 3) show information
prompt_command "${oc_client} -n ${namespace} get pods"
prompt_command "${oc_client} -n ${namespace} get service"
prompt_command "${oc_client} -n ${namespace} describe tangserver"

# Step 4) clevis bind
URL=$(${oc_client} -n "${namespace}" describe tangserver | grep 'Service External URL:' | awk -F 'Service External URL: ' '{print $2}' | sed -e s@/adv@@g | tr -d ' ')
prompt_command_vm_manually "sudo clevis luks bind -f -d ${device} tang '{\"url\":\"${URL}\"}'" "${clevis_ip}"

# Step 5) in json format
prompt_command "${oc_client} -n ${namespace} get tangserver -o json"

# Step 6) rotate
prompt_command "vi ${project_root}/operator_configs/minimal-keyretrieve-rotate/daemons_v1alpha1_tangserver.yaml"
prompt_command "${oc_client} apply -f ${project_root}/operator_configs/minimal-keyretrieve-rotate"

counter=0
while [ $counter -ne "${wait_secs}" ]; do
  ((counter++))
  echo -n -e "\rSleeping until new active keys available ... ${counter}/${wait_secs}"
  sleep 1
done
echo
prompt_command "${oc_client} -n ${namespace} describe tangserver"

# Step 7) clevis rebind
prompt_command_vm_manually "sudo clevis luks report -d ${device} -s 1 -r -q" "${clevis_ip}"

# Step 8) delete
prompt_command "cat ${project_root}/operator_configs/minimal-keyretrieve-deletehiddenkeys/daemons_v1alpha1_tangserver.yaml"
prompt_command "${oc_client} apply -f ${project_root}/operator_configs/minimal-keyretrieve-deletehiddenkeys"
sleep 3
prompt_command "${oc_client} -n ${namespace} describe tangserver"

