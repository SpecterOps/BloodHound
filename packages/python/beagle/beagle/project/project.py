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

import importlib
import os
import pickle
import shutil
import sys

from beagle.aws import s3_write, s3_read, s3_exists
from beagle.facts import GOLANG_WORKSPACE_FILE, COVERAGE_MANIFEST, BHCE_REPO_DIRECTORY
from beagle.semver import Version
from beagle.util import git_version, git_checkout_hash

from typing import List, Dict, Optional


class WorkspaceException(Exception):
    def __init__(self, message: str) -> None:
        super().__init__(message)


class ProjectCoverage:
    def __init__(self, name: str, coverage: float) -> None:
        self.name = name
        self.coverage = coverage


class CoverageManifest:
    def __init__(self) -> None:
        self.projects: Dict[str, ProjectCoverage] = dict()

    def put(self, name: str, coverage: float) -> None:
        self.projects[name] = ProjectCoverage(name=name, coverage=coverage)

    def get(self, name: str) -> ProjectCoverage:
        coverage = self.projects.get(name)

        if coverage is None:
            coverage = ProjectCoverage(name=name, coverage=0.0)
            self.projects[name] = coverage

        return coverage


def marshal_coverage_manifest(manifest: CoverageManifest) -> bytes:
    return pickle.dumps(manifest)


def unmarshal_coverage_manifest(content: bytes) -> CoverageManifest:
    return pickle.loads(content)


def _find_project_root(foss_only=False) -> str:
    file_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = file_dir

    while project_root != os.path.dirname(project_root):
        parent_dir = os.path.dirname(project_root)

        workspace_file = os.path.join(project_root, GOLANG_WORKSPACE_FILE)
        parent_workspace_file = os.path.join(parent_dir, GOLANG_WORKSPACE_FILE)

        # Since the BloodHound Community codebase contains a GOLANG_WORKSPACE_FILE of its own beagle must that it
        # doesn't miss the case where the BloodHound Community codebase is embedded in the enterprise codebase. Beagle
        # must check a directory above for a pipeline file in this case.
        if os.path.lexists(workspace_file) and (foss_only or not os.path.lexists(parent_workspace_file)):
            return project_root

        project_root = parent_dir

    raise WorkspaceException(f"Unable to find project root. Started searching from {file_dir}.")


def _bhce_path(project_dir: str) -> str:
    bhce_path = os.path.join(project_dir, BHCE_REPO_DIRECTORY)

    if os.path.isdir(bhce_path):
        return bhce_path

    return project_dir


class Environment:
    def __init__(
        self,
        local: bool,
        version: Version,
        checkout_hash: str,
    ) -> None:
        self.local = local
        self.version = version
        self.checkout_hash = checkout_hash


class Runtime:
    def __init__(
        self,
        verbose: bool,
        upload_coverage: bool,
        run_integration_tests: bool,
        do_code_generation: bool,
        targets: List[str],
    ) -> None:
        self.verbose = verbose
        self.upload_coverage = upload_coverage
        self.run_integration_tests = run_integration_tests
        self.do_code_generation = do_code_generation
        self.targets = targets


class Filesystem:
    def __init__(
        self,
        env: Environment,
        project_root: str,
    ) -> None:
        self._project = project_root

        # The work directory default is the current working directory of the invoking shell
        self._work = os.getcwd()
        self._artifacts = os.path.join(self._work, "artifacts")

        if env.local:
            # If this is a local build use the project directory to contain the beagle work directory instead of the
            # current working directory
            self._work = os.path.join(self._project, ".beagle")

            # Place artifacts in a folder local to the project root instead of the work directory
            self._artifacts = os.path.join(self._project, "dist")

        # Exported paths
        self.beagle_extensions = os.path.join(self._project, "beagle")

    def _format_path_parts(self, *paths: str) -> List[str]:
        formatted: List[str] = list()

        for path in paths:
            formatted.append(path.lstrip("./").lstrip("/").rstrip("/"))

        return formatted

    def ensure_dir(self, path) -> None:
        if not os.path.lexists(path):
            os.makedirs(path)

    def project_path(self, *paths: str) -> str:
        return os.path.join(self._project, *self._format_path_parts(*paths))

    def cache_path(self, *paths: str) -> str:
        return os.path.join(self._work, "cache", *self._format_path_parts(*paths))

    def log_path(self, *paths: str) -> str:
        return os.path.join(self._work, "logs", *self._format_path_parts(*paths))

    def artifact_path(self, *paths: str) -> str:
        return os.path.join(self._artifacts, *self._format_path_parts(*paths))

    def previous_artifact_path(self, *paths: str) -> str:
        return os.path.join(self._work, "previous_artifacts", *self._format_path_parts(*paths))


class ProjectContext:
    def __init__(
        self,
        env: Environment,
        runtime: Runtime,
        fs: Filesystem,
    ) -> None:
        self.env = env
        self.runtime = runtime
        self.fs = fs

    def print(self, *values, sep=" ", end="\n", file=None):
        print(*values, sep=sep, end=end, file=file)

    def error(self, *values, sep=" ", end="\n"):
        self.print(*values, sep=sep, end=end, file=sys.stderr)

    def info(self, *values, sep=" ", end="\n"):
        if self.runtime.verbose:
            self.print(*values, sep=sep, end=end)

    def load_beagle_extensions(self) -> None:
        if os.path.isdir(self.fs.beagle_extensions):
            sys.path.append(self.fs.beagle_extensions)

            extensions = importlib.import_module("extensions")
            extensions.init(project_ctx=self)

    def setup(self) -> None:
        if not self.runtime.do_code_generation:
            self.info("Code generation steps are disabled")

        self.fs.ensure_dir(self.fs.cache_path())
        self.fs.ensure_dir(self.fs.log_path())
        self.fs.ensure_dir(self.fs.artifact_path())

        if os.path.lexists(self.fs.previous_artifact_path()):
            self.info(f"Previous artifacts found at: {self.fs.previous_artifact_path()}")

            for previous_artifact in os.listdir(self.fs.previous_artifact_path()):
                if os.path.isdir(previous_artifact):
                    shutil.copytree(
                        src=self.fs.previous_artifact_path(previous_artifact),
                        dst=self.fs.artifact_path(previous_artifact),
                    )
                else:
                    shutil.copy(
                        src=self.fs.previous_artifact_path(previous_artifact),
                        dst=self.fs.artifact_path(previous_artifact),
                    )

    def teardown(self) -> None:
        pass

    def read_coverage_manifest(self) -> CoverageManifest:
        try:
            if self.env.local:
                coverage_manifest_path = self.fs.cache_path(COVERAGE_MANIFEST)

                if os.path.lexists(coverage_manifest_path):
                    with open(coverage_manifest_path, "rb") as fin:
                        return unmarshal_coverage_manifest(fin.read())
            elif s3_exists(COVERAGE_MANIFEST):
                return unmarshal_coverage_manifest(s3_read(COVERAGE_MANIFEST))
        except ModuleNotFoundError as ex:
            self.error(f"Unable to unmarshal coverage manifest: {ex.msg}.")

            if self.runtime.upload_coverage:
                self.print("Upstream coverage manifest will be overwritten upon completion of tests.")

        return CoverageManifest()

    def write_coverage_manifest(self, coverage: CoverageManifest) -> None:
        if self.env.local:
            with open(self.fs.cache_path(COVERAGE_MANIFEST), "wb") as fout:
                fout.write(marshal_coverage_manifest(coverage))
        elif self.runtime.upload_coverage:
            s3_write(COVERAGE_MANIFEST, marshal_coverage_manifest(coverage))


def new_project_context(
    local_build: bool,
    verbose: bool,
    run_integration_tests: bool,
    upload_coverage: bool,
    targets: List[str],
    foss_only: bool,
    do_code_generation: bool,
) -> ProjectContext:
    project_root = _find_project_root(foss_only=foss_only)
    env_version_str = os.environ.get("VERSION")
    checkout_hash = os.environ.get("CHECKOUT_HASH")

    if env_version_str is None:
        version = git_version(path=project_root, default_label="v0.0.0")
    else:
        # Trim the leading 'v' if it is present
        if env_version_str.startswith("v"):
            env_version_str = env_version_str[1:]

        version = Version.parse(version=env_version_str)

    if checkout_hash is None:
        checkout_hash = git_checkout_hash(path=project_root)

    env = Environment(
        local=local_build,
        checkout_hash=checkout_hash,
        version=version,
    )

    return ProjectContext(
        env=env,
        runtime=Runtime(
            run_integration_tests=run_integration_tests,
            upload_coverage=upload_coverage,
            verbose=verbose,
            targets=targets,
            do_code_generation=do_code_generation,
        ),
        fs=Filesystem(
            env=env,
            project_root=project_root,
        ),
    )
