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

import os
import platform
import re

from typing import List, Optional

from beagle.plan import Action, BuildPlan, TestPlan, TestException
from beagle.project import ProjectContext
from beagle.subprocess import run_logged, run_simple, run, new_threadpool, as_completed, ThreadPoolExecutor
from beagle.util import json_multi_loads


def golang_os(override: str = "") -> str:
    if override != "":
        return override

    # Attempt to detect the platform
    local_platform = platform.system()

    if local_platform == "Darwin":
        return "darwin"
    elif local_platform == "Linux":
        return "linux"
    elif local_platform == "Windows":
        return "windows"

    raise Exception(f"Unsupported platform: {local_platform}")


# TODO: fill this out for x86/M1/M2 macos if need be
def golang_arch(override: str = "") -> str:
    if override != "":
        return override

    # Attempt to detect the machine
    machine = platform.machine()

    if machine == "x86_64":
        return "amd64"

    raise Exception(f"Unsupported platform: {machine}")


def golang_ldflags(project_ctx: ProjectContext) -> str:
    """
    golang_ldflags looks up the tags for the git repository for the given project context and returns a formatted string
    of ldflags that may be passed as the argument for the go build compiler flag '-ldflags'

    :param project_ctx:
    :return: a formatted string of ldflags
    """
    repo_version = project_ctx.env.version
    ldflags = [
        f"-X 'github.com/specterops/bloodhound/src/version.majorVersion={repo_version.major}'",
        f"-X 'github.com/specterops/bloodhound/src/version.minorVersion={repo_version.minor}'",
        f"-X 'github.com/specterops/bloodhound/src/version.patchVersion={repo_version.patch}'",
    ]

    # If there's no prerelease like '-rc1' then don't bother setting the ldflag for it
    if repo_version.prerelease is not None:
        ldflags.append(
            f"-X 'github.com/specterops/bloodhound/src/version.prereleaseVersion={repo_version.prerelease}'"
        )

    return " ".join(ldflags)


class GoModule:
    def __init__(self, name: str, module: str, path: str, files: Optional[List[str]] = None) -> None:
        self.name = name
        self.module = module
        self.path = path
        self.files = files if files is not None else list()

    @property
    def artifact_name(self) -> str:
        return self.module.split("/")[-1]

    def contains_main_definition(self) -> bool:
        if self.name == "main":
            # If this module has the name "main" then iterate through its files to attempt to find an exported,
            # named function for the entry-point
            for relative_file_path in self.files:
                file_path = os.path.join(self.path, relative_file_path)

                with open(file_path, "r") as fin:
                    for line in fin.readlines():
                        if line.startswith("func main"):
                            return True

        return False

    def requires_code_generation(self) -> bool:
        # Do not bother with anything that's a potential mocked module
        if self.name in ["mock"]:
            return False

        # Look for golang generation directives with grep
        return run_simple(cmd=["grep", "-qR", "go:generate", "./"], cwd=self.path) == 0

    @classmethod
    def list(cls, path: str, recursive: bool = True) -> "List[GoModule]":
        modules: List[GoModule] = list()

        # If not recursive, only list the modules in this directory
        path_target = "./..." if recursive else "./"
        module_list_output = run(cmd=["go", "list", "-json", path_target], cwd=path, capture_stderr=True)

        for module_json in json_multi_loads(module_list_output):
            go_files = module_json.get("GoFiles")

            if go_files is None:
                print(f"Module {module_json['Name']} does not repot having any go files. Skipping.")
                continue

            if isinstance(module_json, list):
                raise Exception("Unexpected type during JSON deserialization: expected a dict but got a list.")

            modules.append(
                GoModule(
                    name=module_json["Name"],
                    module=module_json["ImportPath"],
                    path=module_json["Dir"],
                    files=module_json["GoFiles"],
                )
            )

        return modules

    @classmethod
    def main_modules(cls, path: str) -> "List[GoModule]":
        modules: List[GoModule] = list()

        for module in cls.list(path):
            if module.contains_main_definition():
                modules.append(module)

        return modules


def go_generate(
    project_ctx: ProjectContext, executor: ThreadPoolExecutor, root_module: GoModule, log_path: str
) -> None:
    if not root_module.requires_code_generation():
        return

    project_ctx.info(f"Running code generation for golang module {root_module.module}")

    # Setup environment variables
    environment = os.environ.copy()

    # Make sure GOOS and GOARCH always matches the localhost
    environment.pop("GOOS", None)
    environment.pop("GOARCH", None)

    if not project_ctx.env.local:
        # Set GOPATH since this isn't an engineer's local environment
        environment["GOPATH"] = project_ctx.fs.cache_path("go")

    # Prepare a closure to make orchestrating 'go generate' a little easier
    def _generate(module: GoModule) -> None:
        run_logged(
            cmd=["go", "generate", "./"],
            cwd=module.path,
            environment=environment,
            log_path=log_path,
            print_output=project_ctx.runtime.verbose,
        )

    # Prep the go submodule list
    modules = {module for module in GoModule.list(path=root_module.path) if module.requires_code_generation()}

    # Run through all modules in parallel
    for future in as_completed({executor.submit(_generate, module) for module in modules}):
        future.result()


def sync_workspace(path: str, log_path: str, verbose: bool) -> None:
    run_logged(
        cmd=["go", "work", "sync"],
        cwd=path,
        log_path=log_path,
        print_output=verbose,
    )


class GoWorkspaceException(Exception):
    pass


def _workspace_module_paths(project_ctx: ProjectContext) -> List[str]:
    golang_workspace_path = project_ctx.fs.project_path("go.work")

    if not os.path.exists(path=golang_workspace_path):
        raise GoWorkspaceException(f"Unable to find a valid golang workspace at path: {golang_workspace_path}")

    with open(golang_workspace_path, "r") as fin:
        content_lines = fin.readlines()

    module_paths: List[str] = list()
    in_use_statement = False

    for line in content_lines:
        stripped_line = line.strip()

        if len(stripped_line) == 0 or stripped_line.startswith("//"):
            continue

        if in_use_statement:
            if stripped_line.endswith(")"):
                break

            module_paths.append(stripped_line)
        else:
            if stripped_line.startswith("use"):
                in_use_statement = True

    return module_paths


def workspace_modules(project_ctx: ProjectContext) -> List[GoModule]:
    modules: List[GoModule] = list()

    for module_relative_path in _workspace_module_paths(project_ctx=project_ctx):
        module_path = project_ctx.fs.project_path(module_relative_path)
        module_file_path = project_ctx.fs.project_path(module_relative_path, "go.mod")

        if not os.path.exists(module_file_path):
            raise GoWorkspaceException(f"Unable to find a valid golang module at path: {module_path}")

        with open(module_file_path, "r") as fin:
            content_lines = fin.readlines()

        has_module_def_line = False

        for line in content_lines:
            stripped_line = line.strip()

            if len(stripped_line) == 0 or stripped_line.startswith("//"):
                continue

            if stripped_line.startswith("module"):
                has_module_def_line = True

                # Parse the module details
                module_line_parts = stripped_line.split(" ", 2)

                if len(module_line_parts) != 2:
                    raise GoWorkspaceException(f"Module line is malformed for module file at path: {module_file_path}")

                module_def = module_line_parts[1]
                module_def_parts = module_def.split("/")
                module_name = module_def_parts[-1]

                modules.append(
                    GoModule(
                        name=module_name,
                        module=module_def,
                        path=module_path,
                    )
                )

        if not has_module_def_line:
            raise GoWorkspaceException(
                f"Unable to find golang module details from module file at path: {module_file_path}"
            )

    return modules


class GolangWorkspaceTestPlan(TestPlan):
    def __init__(self, name: str, project_ctx: ProjectContext):
        super().__init__(name=name, source_path=project_ctx.fs.project_path(), project_ctx=project_ctx)

        self._workspace_modules = workspace_modules(project_ctx=self.project_ctx)

    def prepare(self) -> None:
        pass

    def cleanup(self) -> None:
        pass

    def _module_coverage_path(self, module: GoModule) -> str:
        return self.project_ctx.fs.cache_path(f"{module.name}.coverage")

    def _module_log_path(self, module: GoModule) -> str:
        return self.project_ctx.fs.cache_path(f"{module.name}_test.log")

    def _run_module_tests(self, module: GoModule) -> None:
        cmd = [
            "go",
            "test",
            "-ldflags",
            golang_ldflags(project_ctx=self.project_ctx),
            "-coverprofile",
            self._module_coverage_path(module=module),
        ]

        # Integration tests must run sequentially
        if self.project_ctx.runtime.run_integration_tests:
            cmd.append("-p")
            cmd.append("1")
            cmd.append("-tags")
            cmd.append("integration serial_integration")

        # Run all go tests
        cmd.append("./...")

        # Setup environment variables
        environment = os.environ.copy()

        if not self.project_ctx.env.local:
            # Set GOPATH since this isn't an engineer's local environment
            environment["GOPATH"] = self.project_ctx.fs.cache_path("go")

        run_logged(
            cmd=cmd,
            cwd=module.path,
            environment=environment,
            log_path=self._module_log_path(module=module),
            print_output=self.project_ctx.runtime.verbose,
        )

    def run_tests(self) -> None:
        for module in self._workspace_modules:
            self._run_module_tests(module=module)

    def _fetch_module_coverage(self, module: GoModule) -> float:
        output = run_logged(
            cmd=["go", "tool", "cover", "-func", self._module_coverage_path(module=module)],
            cwd=self.source_path,
            log_path=self.project_ctx.fs.log_path(self.name),
        )

        for line in output:
            decoded_line = line.strip()

            if "total:" in decoded_line:
                split_line = re.split("\\s+", decoded_line, maxsplit=3)
                return float(split_line[2].rstrip("%"))

        raise TestException(f"Unable to fetch coverage from go tool.")

    def fetch_coverage(self) -> float:
        num_modules: int = 0
        total_coverage: float = 0.0

        for module in self._workspace_modules:
            total_coverage += self._fetch_module_coverage(module=module)
            num_modules += 1

        return total_coverage / num_modules


class GolangWorkspaceBuildPlan(BuildPlan):
    def __init__(self, name: str, project_ctx: ProjectContext, prepare_actions: Optional[List[Action]] = None) -> None:
        super().__init__(name=name, project_ctx=project_ctx)

        self._workspace_modules = workspace_modules(project_ctx=self.project_ctx)
        self._prepare_actions = prepare_actions if prepare_actions is not None else list()

    def prepare(self) -> None:
        sync_workspace(
            path=self.project_ctx.fs.project_path(),
            log_path=self.project_ctx.fs.log_path(self.name),
            verbose=self.project_ctx.runtime.verbose,
        )

        if self.project_ctx.runtime.do_code_generation:
            with new_threadpool() as executor:
                for module in self._workspace_modules:
                    go_generate(
                        project_ctx=self.project_ctx,
                        executor=executor,
                        root_module=module,
                        log_path=self.project_ctx.fs.log_path(self.name),
                    )

        # Run any additional actions
        for prepare_action in self._prepare_actions:
            prepare_action.run(plan_name=self.name, project_ctx=self.project_ctx)

    def cleanup(self) -> None:
        pass

    def build(self) -> None:
        # Setup environment variables
        environment = os.environ.copy()

        if not self.project_ctx.env.local:
            # Set GOPATH since this isn't an engineer's local environment
            environment["GOPATH"] = self.project_ctx.fs.cache_path("go")

        if environment.get("CGO_ENABLED") is None:
            print("Defaulting to CGO_ENABLED=0")
            environment["CGO_ENABLED"] = "0"

        for root_module in self._workspace_modules:
            for module in GoModule.main_modules(root_module.path):
                artifact_path = self.artifact_path(artifact_name=module.artifact_name)
                self.project_ctx.info(f"Building go binary {module.path} as {artifact_path}")

                run_logged(
                    cmd=["go", "build", "-ldflags", golang_ldflags(project_ctx=self.project_ctx), "-o", artifact_path],
                    cwd=module.path,
                    environment=environment,
                    log_path=self.project_ctx.fs.log_path(self.name),
                    print_output=self.project_ctx.runtime.verbose,
                )
