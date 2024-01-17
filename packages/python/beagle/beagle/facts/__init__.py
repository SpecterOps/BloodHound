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

CI_PIPELINE_FILE = "pipeline.yaml"

GOLANG_WORKSPACE_FILE = "go.work"
GOLANG_COVERAGE_FILE = "coverage.out"

COVERAGE_MANIFEST = "coverage.pyp"

ENV_CACHE_DIR = "CACHE_DIR"
ENV_ARTIFACTS_DIR = "ARTIFACTS_DIR"
ENV_PREVIOUS_ARTIFACTS_DIR = "PREVIOUS_ARTIFACTS_DIR"

CI_S3_BUCKET = "concourse-bhe-build"
BHCE_REPO_DIRECTORY = "bhce"

PRERELEASE_PREFIX = "rc"

GIT_KEEP_FILES = ["keep", ".keep"]
