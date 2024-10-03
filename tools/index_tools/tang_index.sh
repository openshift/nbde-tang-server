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
#   Parameters:
#   1 - $1 image: quay.io/<repository>
#   2 - $2 is for release, format like 0.0.25-1
#   Example:
#   
#   ./tang_index.sh quay.io/sec-eng-special/nbde-tang-server-bundle:1.0.0-0 1.0.0-0
#
#
test -z "${CONTAINER_MGR}" && CONTAINER_MGR='podman'

usage() {
    echo ' ./tang_index.sh quay.io/sec-eng-special/nbde-tang-server-bundle:1.0.0-0 1.0.0-0'
    exit "$2"
}

while getopts "h" arg
do
  case "${arg}" in
    h) usage "$0" 0
       ;;
    *) usage "$0" 1
       ;;
  esac
done

CO_image_with_digest=$1
version=$2

digest=$(echo "${CO_image_with_digest}" | awk -F: '{print $2}')
sub_digest="${digest:0:12}"

test -z "${DO_NOT_LOGIN}" && {
    echo "Login into quay.io ..."
    "${CONTAINER_MGR}" login quay.io -u sarroutb
}
echo "sub_digest:${sub_digest}"

#1. Mirror bundle container image
echo -e "step 1 \n\n"

oc image mirror --filter-by-os=".*" --keep-manifest-list "${CO_image_with_digest}" quay.io/sec-eng-special/nbde-tang-server-bundle-container:"${sub_digest}"
oc image mirror --filter-by-os=".*" --keep-manifest-list "${CO_image_with_digest}" quay.io/sec-eng-special/nbde-tang-server-bundle-container:latest

#2.  Build index image
echo -e "step 2 \n\n"
opm index add --bundles "${CO_image_with_digest}" --tag quay.io/sec-eng-special/nbde-tang-server-index:v"${version}" -c "${CONTAINER_MGR}"

#3. Push image index to quay.io/sec-eng-special/
echo -e "step 3 \n\n"
echo -e "${CONTAINER_MGR} push quay.io/sec-eng-special/nbde-tang-server-index:v$version \n"
${CONTAINER_MGR} push quay.io/sec-eng-special/nbde-tang-server-index:v"${version}"

#4. Substitute version in catalog source
sed -e "s@VERSION_HERE@v${version}@g" < tang_catalog_source_template.yaml > tang_catalog_source.yaml
