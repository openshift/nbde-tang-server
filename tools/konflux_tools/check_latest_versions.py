"""
    This program receives a directory through --directory parameter (.tekton by default)
    and checks the different versions of the images for each of the tools in the directory.
    It then compares the versions with the latest versions available in Quay and prints
    the results or changes the versions in the pipelines directory if the --update flag is set.
    To check which is the current image, a line like the following should be present in the file:
    value: quay.io/konflux-ci/tekton-catalog/task-fbc-related-image-check:0.2@sha256:7a56...a6f
    By default, the program will check quay.io/konflux-ci/tekton-catalog/ as the pattern to check.
    If a different pattern is needed, it can be passed through the --pattern-line parameter.
    The program will return the latest version in the container repository and will update the
    images in case --update parameter is provided.
    The program requires next tools:
    - skopeo
    - sed
    - grep
    - python3
"""
import argparse
import json
import subprocess
from os import listdir
from os.path import isfile, join

DEFAULT_FILTER_LINE = "quay.io/konflux-ci/tekton-catalog/"
DEFAULT_DIRECTORY = "./.tekton"

class Logger: # pylint: disable=too-few-public-methods
    """ Logger class """
    def __init__(self, verbose):
        self.verbose = verbose

    def vprint(self, *args, sep=" "):
        """ Print only if verbose is set """
        if self.verbose:
            print(*args, sep=sep)

def parse_arguments():
    """ Parse the arguments of the program """
    parser = argparse.ArgumentParser(
        description="Check the latest versions of the tools in the given directory")
    parser.add_argument("--directory", default=DEFAULT_DIRECTORY, help=
                        "The directory to check the versions of the tools " +
                        f"(default: {DEFAULT_DIRECTORY})")
    parser.add_argument("--pattern", help=
                        "The pattern parameter to use to check the versions of the tools")
    parser.add_argument("--file-pattern", help=
                        "File pattern to search for in the directory")
    parser.add_argument("--file-exclude-pattern", help=
                        "File pattern to exclude in directory search")
    parser.add_argument("--image-pattern", help=
                        "Image pattern to apply in image search")
    parser.add_argument("--image-exclude-pattern", help=
                        "Image pattern to exclude in image search")
    parser.add_argument("--update",help=
                        "Update the versions of the tools in the pipelines directory",
                        action="store_true")
    parser.add_argument("--verbose", help=
                        "Print verbose output", action="store_true")
    return parser.parse_args()

class VersionChecker:
    """ Class to check the versions of the tools in the given directory """
    def __init__(self, input_args):
        """ Constructor """
        self.parsed_images = {}
        self.input_args = input_args
        self.logger = Logger(self.input_args.verbose)

    def get_latest_tag_from_list(self, image, tags):
        """ Get the latest tag from a list of tags """
        most_recent_created = "2000-01-01T00:00:00.000000000Z"
        most_recent_tag = ""
        for tag in tags:
            result = subprocess.run(["skopeo", "inspect", "-n", "docker://" + image + ":" + tag],
                                    stdout=subprocess.PIPE, check=False)
            created = json.loads(result.stdout.decode("utf-8"))["Created"]
            self.logger.vprint("Tag:", tag, ", Created:", created, sep="")
            if created >= most_recent_created:
                most_recent_created = created
                most_recent_tag = tag
        self.logger.vprint("Most recent tag:", most_recent_tag)
        return most_recent_tag


    def get_digest_from_tag(self, image_with_tag):
        """ This function gets the digest of the latest tag """
        result = subprocess.run(["skopeo", "inspect", "-n", "docker://" + image_with_tag],
                                stdout=subprocess.PIPE, check=False)
        return json.loads(result.stdout.decode("utf-8"))["Digest"]

    def get_latest_tag(self, image):
        """ Get the latest tag of an image:
            This function extracts all the tags for a given image through skopeo
            and for each tag, it extracts the date and returns the more recent tag
        """
        if image in self.parsed_images:
            self.logger.vprint("WARNING: Already parsed image:", image)
            return self.parsed_images[image]
        result = subprocess.run(["skopeo", "list-tags", "docker://" + image],
                                stdout=subprocess.PIPE, check=False)
        self.logger.vprint("Result:", result.stdout.decode("utf-8"))
        tags = json.loads(result.stdout.decode("utf-8"))["Tags"]
        self.logger.vprint("Tags:", tags)
        # Get latest tag
        latest_tag = self.get_latest_tag_from_list(image, tags)
        digest = self.get_digest_from_tag(image + ":" + latest_tag)
        latest_tag_no_hyphen = latest_tag.split("-")[0]
        ltag = image + ":" + latest_tag_no_hyphen + "@" + digest
        self.parsed_images[image] = ltag
        return ltag

    def update_file(self, file_name, latest_version):
        """ Update the file with the latest version """
        self.logger.vprint("Updating file:", file_name)
        tag_version = latest_version.split(":")[1].split("@")[0]
        if tag_version.find("-") != -1:
            image_with_no_tag = latest_version.split("@")[0].rsplit("-", 1)[0].split(":")[0]
        else:
            image_with_no_tag = latest_version.split("@")[0].split(":")[0]
        cmd = "grep -q " + latest_version + " " + file_name
        self.logger.vprint("Grep command:", cmd)
        r = subprocess.run(cmd, stdout=subprocess.PIPE, shell=True, check=False)
        if r.returncode != 0:
            self.logger.vprint("Updating image:", image_with_no_tag, "to:", latest_version)
            cmd = ("sed -i -E 's!" + image_with_no_tag +
                ":[0-9]{0,1}\\.[0-9]{0,1}@sha256:[0-9,a-f]{6,}!" +
                latest_version + "!g' " + file_name)
            self.logger.vprint("Command:", cmd)
            r = subprocess.run(cmd, stdout=subprocess.PIPE, shell=True, check=False)
            if r.returncode == 0:
                print("Updated file:", file_name, "with version:", latest_version)
            else:
                print("Unexpected error updating file:", file_name, "with version:", latest_version)
        else:
            print("File:", file_name, "does not need updating")

    def image_applies(self, image):
        """ Check if the image applies to the patterns:
            - image_pattern: image must contain this pattern
            - image_exclude_pattern: image must not contain this pattern
        """
        if self.input_args.image_exclude_pattern:
            if self.input_args.image_exclude_pattern in image:
                self.logger.vprint(f"Excluding image:{image} (exclude pattern):" +
                                    f"{self.input_args.image_exclude_pattern}")
                return False
        if self.input_args.image_pattern:
            if self.input_args.image_pattern not in image:
                self.logger.vprint(f"Excluding image:{image} (pattern not found):" +
                                    f"{self.input_args.image_pattern}")
                return False
        return True

    def process_latest_versions(self, file, pattern_filter=DEFAULT_FILTER_LINE):
        """ Get the latest versions of the tools in the file """
        for line in file:
            if pattern_filter in line:
                image = (pattern_filter + line.split(pattern_filter)[1])\
                    .strip(" ").rstrip(" ").strip().rstrip()
                if self.image_applies(image):
                    non_sha_image = image.split("@")[0]
                    non_tag_image = non_sha_image.split(":")[0]
                    self.logger.vprint("Image:", image, sep="")
                    self.logger.vprint("No SHA Image:", non_sha_image, sep="")
                    self.logger.vprint("No tag Image:", non_tag_image, sep="")
                    latest_version = self.get_latest_tag(non_tag_image)
                    print("Latest tag with digest:->", latest_version, "<-", sep="")
                    if self.input_args.update:
                        self.update_file(file.name, latest_version)

    def file_applies(self, file_name):
        """ Check if the file applies to the patterns:
            - file_pattern: file_name must contain this pattern
            - file_exclude_pattern: file_name must not contain this pattern
        """
        if self.input_args.file_exclude_pattern:
            if self.input_args.file_exclude_pattern in file_name:
                self.logger.vprint(f"Excluding file:{file_name} (exclude pattern):" +
                                    f"{self.input_args.file_exclude_pattern}")
                return False
        if self.input_args.file_pattern:
            if self.input_args.file_pattern not in file_name:
                self.logger.vprint(f"Excluding file:{file_name} (pattern not found):" +
                                    f"{self.input_args.file_pattern}")
                return False
            return self.input_args.file_pattern in file_name
        return True

    def read_files(self, directory):
        """ Read all files in directory and return a list with the files """
        onlyfiles = [(directory + "/" + f).strip().replace("//", "/") for f in listdir(directory)
                    if isfile(join(directory, f)) and self.file_applies(f)]
        self.logger.vprint(onlyfiles)
        return onlyfiles

    def directory_to_parse(self):
        """ Get the directory to parse """
        if self.input_args.directory:
            return self.input_args.directory
        return DEFAULT_DIRECTORY

    def check_versions(self):
        """ Read all files in directory """
        files = self.read_files(self.directory_to_parse())
        for file in files:
            with open(file, encoding='utf-8') as f:
                if self.input_args.pattern:
                    self.process_latest_versions(f, self.input_args.pattern)
                else:
                    self.process_latest_versions(f)

def main():
    """ Main method """
    VersionChecker(parse_arguments()).check_versions()

if __name__ == "__main__":
    main()
