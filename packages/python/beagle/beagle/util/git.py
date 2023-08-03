# Copyright 2023 Specter Ops, Inc.
# 
# Licensed under the Apache License, Version 2.0
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# 
# SPDX-License-Identifier: Apache-2.0

import subprocess

from beagle.semver import Version
from beagle.subprocess import run


def _get_latest_version_label(path: str, default_label: str) -> Version:
    """
    Gets the latest version label from the underlying git repository. Versions must start with the prefix 'v' in order
    to be recognized by this function.

    :param path: Path to the git repository
    :param default_label: Default version label to use when no tags can be found
    :return: String representation of the current version label
    """
    output = run(cmd=["git", "tag", "--list", "v*"], cwd=path)

    # Parse out the default version by stripping the version prefix
    default_version = Version.parse(default_label[1:])
    latest_version = default_version

    for line in output.readlines():
        formatted_line = line.strip()

        # Skip empty lines
        if len(formatted_line) == 0:
            continue

        # Parse the next version and compare it against the current latest version
        next_version = Version.parse(formatted_line[1:])

        if next_version > latest_version:
            latest_version = next_version

    return latest_version


def git_version(path: str, default_label: str) -> Version:
    # Get the latest version label from the repo tags
    return _get_latest_version_label(path=path, default_label=default_label)


def git_label_hash(path: str, label: str) -> str:
    output = run(cmd=["git", "show-ref", "-s", label], cwd=path)
    return output.read().strip()


def git_checkout_hash(path: str) -> str:
    output = run(cmd=["git", "rev-list", "--max-count=1", "HEAD"], cwd=path)
    return output.read().strip()


def git_status(path: str) -> str:
    output = run(cmd=["git", "status"], cwd=path)
    return output.read().strip()


def git_path_has_changes(repo_dir: str, path: str) -> bool:
    result = subprocess.run(["git", "-C", repo_dir, "diff", "--quiet", "HEAD", "--", path], stderr=subprocess.PIPE)
    return result.returncode > 0
