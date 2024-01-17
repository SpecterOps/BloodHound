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

from setuptools import setup, find_packages


def readme() -> str:
    with open("README.md", "r") as fin:
        return fin.read()


def requirements() -> str:
    with open("requirements.txt", "r") as fin:
        return fin.read()


setup(
    name="beagle",
    version="0.0.1",
    author="BHE Engineering",
    author_email="bloodhoundenterprise@specterops.io",
    description="Test, Build and CI Automation for the BHE Project Cluster",
    long_description=readme(),
    long_description_content_type="text/markdown",
    url="https://github.com/SpecterOps/bloodhound",
    packages=find_packages(),
    install_requires=requirements(),
    license="Apache-2.0",
    classifiers=[
        "Programming Language :: Python :: 3.10",
        "License :: OSI Approved :: Apache Software License",
    ],
)
