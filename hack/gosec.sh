#!/bin/sh -eux
#
#   Copyright [2024] [sarroutb (at) redhat.com]
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
usage() {
    echo ''
    echo './gosec.sh [-s severity] [-c confidence] [-p path]'
    echo 'Example:'
    echo '         ./gosec.sh -s medium -c medium -p ./... (default)'
    echo ''
    exit "$2"
}
security="medium"
confidence="medium"
path="./..."

while getopts "s:c:p:h" arg
do
  case "${arg}" in
    c) confidence="${OPTARG}"
       ;;
    p) path="${OPTARG}"
       ;;
    s) security="${OPTARG}"
       ;;
    h) usage "$0" 0
       ;;
    *) usage "$0" 1
       ;;
  esac
done

GOFLAGS='' go install github.com/securego/gosec/v2/cmd/gosec@latest
type gosec || error "gosec application not installed"
gosec -severity "${security}" -confidence "${confidence}" "${path}"
