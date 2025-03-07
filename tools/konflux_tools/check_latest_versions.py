""" This program receives a directory through --directory parameter (.tekton by default)
    and checks the different versions of the images for each of the tools in the directory.
    It then compares the versions with the latest versions available in Quay and prints
    the results or changes the versions in the pipelines directory if the --update flag is set.
    To check which is the current image, a line like the following should be present in the file:
    value: quay.io/konflux-ci/tekton-catalog/task-fbc-related-image-check:0.2@sha256:7a56...a6f
    By default, the program will check quay.io/konflux-ci/tekton-catalog/ as the pattern to check.
    If a different pattern is needed, it can be passed through the --pattern-line parameter.
    The program will check return the latest version in the container repository """
import argparse
import json
import subprocess
from os import listdir
from os.path import isfile, join

FILTER_LINE = "quay.io/konflux-ci/tekton-catalog/"
verbose = False
parsed_images = {}

def vprint(*args, sep=" "):
    """ Print only if verbose is set """
    global verbose
    if verbose:
        print(*args, sep=sep)

def parse_arguments():
    """ Parse the arguments of the program """
    parser = argparse.ArgumentParser(
        description="Check the latest versions of the tools in the given directory")
    parser.add_argument("--directory", help=
                        "The directory to check the versions of the tools")
    parser.add_argument("--pattern", help=
                        "The pattern parameter to use to check the versions of the tools")
    parser.add_argument("--update",help=
                        "Update the versions of the tools in the pipelines directory",
                        action="store_true")
    parser.add_argument("--verbose", help=
                        "Print verbose output", action="store_true")
    return parser.parse_args()

# This function returs a list of all the files inside a directory
def read_files(directory):
    """ Read all files in directory """
    onlyfiles = [(directory + "/" + f).strip() for f in listdir(directory)
                 if isfile(join(directory, f))]
    vprint(onlyfiles)
    return onlyfiles

# This function gets a list of tags for a given image
# and returns the most recent tag
def get_latest_tag_from_list(image, tags):
    """ Get the latest tag from a list of tags """
    most_recent_created = "2000-01-01T00:00:00.000000000Z"
    most_recent_tag = ""
    for tag in tags:
        result = subprocess.run(["skopeo", "inspect", "-n", "docker://" + image + ":" + tag],
                                stdout=subprocess.PIPE, check=False)
        created = json.loads(result.stdout.decode("utf-8"))["Created"]
        vprint("Tag:", tag, ", Created:", created, sep="")
        if created >= most_recent_created:
            most_recent_created = created
            most_recent_tag = tag
    vprint("Most recent tag:", most_recent_tag)
    return most_recent_tag

# This function gets the digest of the latest tag
def get_digest_from_tag(image_with_tag):
    """ Get digest from tag """
    result = subprocess.run(["skopeo", "inspect", "-n", "docker://" + image_with_tag],
                            stdout=subprocess.PIPE, check=False)
    return json.loads(result.stdout.decode("utf-8"))["Digest"]

# This function extracts all the tags for a given image through skopeo
# and for each tag, it extracts the date and returns the more recent tag
def print_latest_tag(image):
    """ Get the latest tag of an image """
    global parsed_images
    if image in parsed_images:
        vprint("WARNING: Already parsed image:", image)
        parsed_images[image] = True
        return
    result = subprocess.run(["skopeo", "list-tags", "docker://" + image], stdout=subprocess.PIPE,
                            check=False)
    vprint("Result:", result.stdout.decode("utf-8"))
    tags = json.loads(result.stdout.decode("utf-8"))["Tags"]
    vprint("Tags:", tags)
    # Get latest tag
    latest_tag = get_latest_tag_from_list(image, tags)
    digest = get_digest_from_tag(image + ":" + latest_tag)
    print("Latest tag with digest:->", image + ":" + latest_tag + "@" + digest, "<-")

def print_latest_versions(file, pattern_filter=FILTER_LINE):
    """ Get the latest versions of the tools in the file """
    for line in file:
        if pattern_filter in line:
            # Remove the value: part by splitting the line from filter
            image = (pattern_filter + line.split(pattern_filter)[1]).strip(" ").rstrip(" ")
            non_sha_image = image.split("@")[0]
            non_tag_image = non_sha_image.split(":")[0]
            vprint("Image:", image.strip().rstrip(), sep="")
            vprint("No SHA Image:", non_sha_image, sep="")
            vprint("No tag Image:", non_tag_image, sep="")
            print_latest_tag(non_tag_image)


def check_versions(input_args):
    """ Read all files in directory """
    files = read_files(input_args.directory)
    for file in files:
        with open(file, encoding='utf-8') as f:
            if input_args.pattern:
                print_latest_versions(f, input_args.pattern)
            else:
                print_latest_versions(f)

def main():
    """ Main method """
    input_args = parse_arguments()
    global verbose
    verbose = input_args.verbose
    check_versions(input_args)

if __name__ == "__main__":
    main()
