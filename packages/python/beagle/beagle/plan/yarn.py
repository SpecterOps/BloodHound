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

import json
import os
import pathlib

from beagle.plan import TestPlan, BuildPlan
from beagle.project import ProjectContext
from beagle.subprocess import run_logged, run


class YarnTestPlan(TestPlan):
    def __init__(
        self,
        name: str,
        source_path: str,
        project_ctx: ProjectContext,
        yarn_workspace = False,
    ) -> None:
        super().__init__(
            name=name,
            source_path=source_path,
            project_ctx=project_ctx,
        )

        self.source_path = source_path
        self.yarn_workspace = yarn_workspace,
        self.node_modules_path = os.path.join(self.source_path, "node_modules")

    def prepare(self) -> None:
        run(
            cmd=["yarn", "install"],
            cwd=self.source_path,
            capture_stderr=True,
        )

    def cleanup(self) -> None:
        pass

    def run_tests(self) -> None:
        run_logged(
            cmd=["yarn", "test", "--coverage", "--run"],
            cwd=self.source_path,
            log_path=self.project_ctx.fs.log_path(self.name),
            print_output=self.project_ctx.runtime.verbose,
        )

    def fetch_coverage(self) -> float:
        coverage_summary_path = os.path.join(self.source_path, "coverage", "coverage-summary.json")

        if not os.path.exists(coverage_summary_path):
            return 0

        with open(coverage_summary_path, "r") as fin:
            coverage_summary = json.load(fin)

        return coverage_summary["total"]["lines"]["pct"]


class YarnBuildPlan(BuildPlan):
    def __init__(self, name: str, source_path: str, project_ctx: ProjectContext) -> None:
        super().__init__(name, project_ctx)

        self.source_path = source_path
        self.node_modules_path = os.path.join(self.source_path, "node_modules")

    def prepare(self) -> None:
        run(
            cmd=["yarn", "install"],
            cwd=self.source_path,
            capture_stderr=True,
        )

    def cleanup(self) -> None:
        pass

    def build(self):
        # Find and ensure that the artifact path is prepared
        artifacts_path = self.artifact_path(self.name)
        self.project_ctx.fs.ensure_dir(artifacts_path)

        # Set the build path environment variable to output to the prepared artifact directory
        environment = os.environ.copy()
        environment["BUILD_PATH"] = artifacts_path

        run_logged(
            cmd=["yarn", "build"],
            cwd=self.source_path,
            log_path=self.project_ctx.fs.log_path(self.name),
            print_output=self.project_ctx.runtime.verbose,
            environment=environment,
        )


class UIBuildPlan(BuildPlan):
    def __init__(self, name: str, source_path: str, project_ctx: ProjectContext) -> None:
        super().__init__(name, project_ctx)

        self.source_path = source_path
        self.node_modules_path = os.path.join(self.source_path, "node_modules")

    def prepare(self) -> None:
        run(
            cmd=["yarn", "install"],
            cwd=os.path.join(
                pathlib.Path(__file__).parent.resolve(), "..", "..", "..", "..", "javascript", "bh-shared-ui"
            ),
            capture_stderr=True,
        )
        run_logged(
            cmd=["yarn", "build"],
            cwd=os.path.join(
                pathlib.Path(__file__).parent.resolve(), "..", "..", "..", "..", "javascript", "bh-shared-ui"
            ),
            log_path=self.project_ctx.fs.log_path(self.name),
            print_output=self.project_ctx.runtime.verbose,
        )
        run(
            cmd=["yarn", "install"],
            cwd=os.path.join(
                pathlib.Path(__file__).parent.resolve(), "..", "..", "..", "..", "javascript", "js-client-library"
            ),
            capture_stderr=True,
        )
        run_logged(
            cmd=["yarn", "build"],
            cwd=os.path.join(
                pathlib.Path(__file__).parent.resolve(), "..", "..", "..", "..", "javascript", "js-client-library"
            ),
            log_path=self.project_ctx.fs.log_path(self.name),
            print_output=self.project_ctx.runtime.verbose,
        )
        run(
            cmd=["yarn", "install"],
            cwd=self.source_path,
            capture_stderr=True,
        )

    def cleanup(self) -> None:
        pass

    def build(self):
        # Find and ensure that the artifact path is prepared
        artifacts_path = self.artifact_path(self.name)
        self.project_ctx.fs.ensure_dir(artifacts_path)

        # Set the build path environment variable to output to the prepared artifact directory
        environment = os.environ.copy()
        environment["BUILD_PATH"] = artifacts_path

        run_logged(
            cmd=["yarn", "build"],
            cwd=self.source_path,
            log_path=self.project_ctx.fs.log_path(self.name),
            print_output=self.project_ctx.runtime.verbose,
            environment=environment,
        )
