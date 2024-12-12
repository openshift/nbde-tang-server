#!/usr/bin/env python3
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

    IMPORTANT:  This program assumes access to cluster node is granted and it is possible
                to run commands (podman) in it.
    IMPORTANT:  This program is not complete. It only parses input parameters and shows them.
"""
import argparse

def show_parameters(args):
    """
    Debug method to show parsed parameters
    """
    print("Original image name:", args.original_image)
    print("Cluster Password:", args.cluster_password)
    if args.cluster_name:
        print("Cluster Name:", args.cluster_name)

def parse_parameters():
    """
    This function initiates next input parameters
    - Original image name (--original_image)
    - Cluster Password (--cluster_password)
    - Cluster Name (optional, it will be guessed if not provided) (--cluster_name)
    """
    parser = argparse.ArgumentParser(prog='mirror_image.py',
        description='Mirror image from an initiating registry to the cluster local mirror')
    parser.add_argument('-o', '--original_image', help='Original image name',
        required=True)
    parser.add_argument('-p', '--cluster_password', help='Cluster Password',
        required=True)
    parser.add_argument('-n', '--cluster_name', help='Cluster Name (guessed if not provided)',
        required=False)
    return parser.parse_args()

def main():
    """
    Main method:
    - parse parameters
    - perform mirroring
    """
    show_parameters(parse_parameters())

if __name__ == '__main__':
    main()
