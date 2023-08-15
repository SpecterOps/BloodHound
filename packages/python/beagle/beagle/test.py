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
import time

import plans

from beagle.facts import BHCE_REPO_DIRECTORY
from beagle.plan import TestPlan
from beagle.project import ProjectContext, CoverageManifest
from beagle.subprocess import run
from beagle.util import git_path_has_changes


def _gate_coverage(test_plan: TestPlan, project_ctx: ProjectContext, coverage: CoverageManifest) -> None:
    previous_coverage = coverage.get(test_plan.name)
    current_coverage = test_plan.fetch_coverage()

    # Validate coverage within a half-percent margin
    if current_coverage - previous_coverage.coverage <= -1.5:
        print(
            f"Test coverage {current_coverage}% for test plan {test_plan.name} is less than previously reported "
            f"coverage of: {previous_coverage.coverage}%."
        )

        if not project_ctx.env.local:
            raise SystemExit(1)
    else:
        coverage.put(name=test_plan.name, coverage=current_coverage)

    print(f"Total coverage for test plan {test_plan.name}: {current_coverage:.2f}%")


def _run_test_plan(test_plan: TestPlan, project_ctx: ProjectContext, coverage: CoverageManifest) -> None:
    project_ctx.info(f"Executing test plan: {test_plan.name}")

    try:
        test_plan.prepare()
        test_plan.run_tests()
    finally:
        test_plan.cleanup()

    _gate_coverage(test_plan=test_plan, project_ctx=project_ctx, coverage=coverage)

    # Output a new line to break output between plans
    project_ctx.info()


def _start_integration_test_services(project_ctx: ProjectContext) -> None:
    project_ctx.info("Starting integration test services")

    run(cmd=["service", "neo4j", "start"], cwd=os.getcwd())
    run(cmd=["service", "postgresql", "start"], cwd=os.getcwd())

    # Wait 15 seconds for the services to online
    time.sleep(15)

def _stop_integration_test_services(project_ctx: ProjectContext) -> None:
    project_ctx.info("Stopping integration test services")

    run(cmd=["service", "neo4j", "stop"], cwd=os.getcwd())
    run(cmd=["service", "postgresql", "stop"], cwd=os.getcwd())


def run_tests(project_ctx: ProjectContext, targeted: bool) -> None:
    if not project_ctx.env.local:
        _start_integration_test_services(project_ctx)

    coverage = project_ctx.read_coverage_manifest()

    for test_plan in plans.test_plans:
        if len(project_ctx.runtime.targets) > 0 and test_plan.name not in project_ctx.runtime.targets:
            continue

        # Check to see if we need to check the BloodHound Community repository for changes
        repo_dir = project_ctx.fs.project_path()

        if BHCE_REPO_DIRECTORY in test_plan.source_path:
            repo_dir = os.path.join(project_ctx.fs.project_path(), BHCE_REPO_DIRECTORY)

        if targeted and not git_path_has_changes(repo_dir=repo_dir, path=test_plan.source_path):
            project_ctx.print(
                f"No changes detected in {test_plan.source_path}, skipping tests. "
                "To force a run use the '-a' or '--all-tests' flag."
            )
            continue

        _run_test_plan(test_plan=test_plan, project_ctx=project_ctx, coverage=coverage)

    project_ctx.write_coverage_manifest(coverage=coverage)

    if not project_ctx.env.local:
        _stop_integration_test_services(project_ctx)
