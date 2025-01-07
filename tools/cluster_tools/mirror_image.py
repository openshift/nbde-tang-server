#!/usr/bin/env python3
#
# Copyright 2024 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
"""
This Python program performs mirroring from an initiating registry to the cluster local mirror
To do so, it follows next steps:
	oc get nodes
	# With previous command, it parses one particular node name
	oc debug nodes/<node_name>
	sh-4.2# chroot /host
	# For next command, password must be provided as an input parameter.
	# Server can be provided as input parameter. If not, it will be tried to be guessed.
	sh-4.2# oc login -u kubeadmin -p <password_from_install_log>
            https://api-int.<cluster_name>.<base_domain>:6443
	sh-4.2# podman login -u kubeadmin -p $(oc whoami -t)
            image-registry.openshift-image-registry.svc:5000
	# For next commands, original image name must be provided as an input parameter
	sh-4.2# podman pull <name.io>/<image>
	sh-4.2# podman tag <name.io>/<image>
            image-registry.openshift-image-registry.svc:5000/openshift/<image>
	sh-4.2# podman push image-registry.openshift-image-registry.svc:5000/openshift/<image>

    Program usage:
    python3 ./tools/cluster_tools/mirror_image.py -h
    usage:  mirror_image.py [-h] -c CLUSTER_PASSWORD -o ORIGINAL_IMAGE
            [-f FILE_FOR_USER_PASSWORD] [-u USER] [-p PASSWORD] [-n CLUSTER_NAME]

    Mirror image from an initiating registry to the cluster local mirror

    options:
    -h, --help            show this help message and exit
    -c CLUSTER_PASSWORD, --cluster_password CLUSTER_PASSWORD
                            Cluster Password
    -o ORIGINAL_IMAGE, --original_image ORIGINAL_IMAGE
                            Original image name
    -f FILE_FOR_USER_PASSWORD, --file-for-user-password FILE_FOR_USER_PASSWORD
                            File to provide user and password, (user:password)
    -u USER, --user USER  User to connect to brew.registry.redhat.io
    -p PASSWORD, --password PASSWORD
                            Password to connect to brew.registry.redhat.io
    -n CLUSTER_NAME, --cluster_name CLUSTER_NAME
                            Cluster Name (guessed if not provided)

    IMPORTANT:  This program assumes access to cluster node is granted and it is possible
                to run commands (podman) in it.
    USER: YOUR USER AND PASSWORD WILL BE STORED IN THE COMMAND HISTORY AND SHOWN IN THE SCREEN
          IF USING USER AND PASSWORD AS INPUT PARAMETERS.
          IT IS RECOMMENDED TO USE FILE (-f) FOR USER AND PASSWORD, WITH NEXT FORMAT:
          user:password
"""
import argparse
import subprocess
import sys
import pexpect

WORKER_LINE = "worker"
NODE_CONNECTION_TIMEOUT = 20
LOCAL_REGISTRY = "image-registry.openshift-image-registry.svc:5000"

def get_line(all_lines, key):
    """
    This function returns the line that contains the key
    """
    for line in all_lines.split('\n'):
        if key in line:
            return line
    return None

def send_pexpect_command(child, command, expected):
    """
    This function sends a command to the child and waits for the expected result
    """
    child.sendline(command)
    child.expect(expected)
    sys.stdout.flush()

def get_image_registry(image):
    """
    This function returns the image registry
    """
    return image.split('/')[0]

def get_image_path(image):
    """
    This function returns the image path
    """
    return 'openshift/' + image.split('/', 1)[1].split('/',1)[1]

def get_user_password_from_file(file_name):
    """
    This function returns the user and password from a file
    """
    with open(file_name, 'r', encoding="utf-8") as file:
        user_password = file.readline().strip()
    return user_password.split(':')

def get_user_password_from_args(args):
    """
    This function returns the user and password from the arguments
    """
    if args.user and args.password:
        return args.user, args.password
    return get_user_password_from_file(args.file_for_user_password)

def launch_node_commands(node, api_server, args):
    """
    This function launches commands in the node
    """
    user, password = get_user_password_from_args(args)
    child = pexpect.spawnu('oc debug nodes/' + node, timeout=NODE_CONNECTION_TIMEOUT)
    child.logfile_read = sys.stdout
    child.expect('# ')
    send_pexpect_command(child, '\n', '#')
    send_pexpect_command(child, 'chroot /host', '#')
    send_pexpect_command(child, '\n', '#')
    send_pexpect_command(child, 'oc login -u kubeadmin -p ' + args.cluster_password +
                         ' ' + api_server, ': ')
    index = child.expect([":", "#", pexpect.TIMEOUT])
    if index == 0:
        send_pexpect_command(child, 'yes', '#')
    else:
        send_pexpect_command(child, '\n', '#')
    send_pexpect_command(child, 'podman login -u kubeadmin -p $(oc whoami -t) ' +
                         LOCAL_REGISTRY, '#')
    send_pexpect_command(child, 'echo', '#')
    child.expect('#')
    send_pexpect_command(child, 'podman login -u ' + user +
                         ' ' + get_image_registry(args.original_image), 'Password')
    send_pexpect_command(child, password, '#')
    send_pexpect_command(child, 'podman pull ' + args.original_image, '#')
    send_pexpect_command(child, 'podman tag ' + args.original_image + ' ' + LOCAL_REGISTRY +
                         '/' + get_image_path(args.original_image), '#')
    send_pexpect_command(child, 'podman images | grep ' + LOCAL_REGISTRY, '#')
    send_pexpect_command(child, 'podman push ' + LOCAL_REGISTRY + '/' + get_image_path(args.original_image), '#')
    send_pexpect_command(child, 'exit', '#')
    sys.stdout.flush()
    child.close()

def execute_command(cmd):
    """
    This function executes command and returs its result
    """
    cmd_ret = subprocess.run(cmd, stdout=subprocess.PIPE, text=True, check=True)
    return cmd_ret.stdout

def get_worker_node_name():
    """
    This function returns the worker node name
    """
    return get_line(execute_command(["oc", "get", "nodes"]), WORKER_LINE).split()[0]

def get_api_server():
    """
    This function returns the API server
    """
    return execute_command(["oc", "whoami", "--show-server"])

def run_command_sequence(args):
    """
    This function runs the command sequence in worker
    """
    node = get_worker_node_name()
    print(node)
    api_server = get_api_server()
    launch_node_commands(node, api_server, args)
    print()

def parse_parameters():
    """
    This function initiates next input parameters
    - Original image name (--original_image)
    - Cluster Password (--cluster_password)
    - Cluster Name (optional, it will be guessed if not provided) (--cluster_name)
    """
    parser = argparse.ArgumentParser(prog='mirror_image.py',
        description='Mirror image from an initiating registry to the cluster local mirror')
    parser.add_argument('-c', '--cluster_password', help='Cluster Password',
                        required=True)
    parser.add_argument('-o', '--original_image', help='Original image name',
        required=True)
    parser.add_argument('-f', '--file-for-user-password',
                        help='File to provide user and password, (user:password)',
                        required=False)
    parser.add_argument('-u', '--user', help='User to connect to brew.registry.redhat.io',
                        required=False)
    parser.add_argument('-p', '--password', help='Password to connect to brew.registry.redhat.io',
                        required=False)
    parser.add_argument('-n', '--cluster_name', help='Cluster Name (guessed if not provided)',
                        required=False)
    return parser.parse_args()

def main():
    """
    Main method:
    - parse parameters
    - perform mirroring
    """
    run_command_sequence(parse_parameters())

if __name__ == '__main__':
    main()
