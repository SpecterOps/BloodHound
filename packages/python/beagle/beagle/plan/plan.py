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

from beagle.project import ProjectContext


class TestException(Exception):
    def __init__(self, msg: str) -> None:
        self.msg = msg

    def __str__(self) -> str:
        return self.msg


class Action:
    def run(self, plan_name: str, project_ctx: ProjectContext) -> None:
        pass


class TestPlan:
    def __init__(
        self,
        name: str,
        source_path: str,
        project_ctx: ProjectContext,
    ):
        self.name = name
        self.source_path = source_path
        self.project_ctx = project_ctx

    def prepare(self) -> None:
        pass

    def cleanup(self) -> None:
        pass

    def run_tests(self) -> None:
        raise NotImplementedError()

    def fetch_coverage(self) -> float:
        raise NotImplementedError()


class BuildPlan:
    def __init__(self, name: str, project_ctx: ProjectContext):
        self.name = name
        self.project_ctx = project_ctx

    def artifact_path(self, artifact_name: str) -> str:
        return self.project_ctx.fs.artifact_path(artifact_name)

    def prepare(self) -> None:
        pass

    def cleanup(self) -> None:
        pass

    def build(self) -> None:
        raise NotImplementedError()

    def execute(self) -> None:
        try:
            self.prepare()
            self.build()
        finally:
            self.cleanup()
