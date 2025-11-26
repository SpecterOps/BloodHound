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

# Import global plans so that beagle extensions can add to a registry of prepared plans available to run
import plans

# Supporting beagle imports
from beagle.plan import (
    GolangWorkspaceBuildPlan,
    GolangWorkspaceTestPlan,
    CopyBloodHoundUIAssets,
    YarnTestPlan,
    YarnBuildPlan,
)
from beagle.project import ProjectContext


def init(project_ctx: ProjectContext) -> None:
    """
    init is the beagle initializer function. This is the entry point for these particular extensions.

    :param project_ctx: the project context contains all context of the beagle environment, runtime and project files
    :return:
    """
    project_ctx.info("FOSS beagle extensions enabled")

    plans.build_plans = [
        YarnBuildPlan(
            name="bh-ui",
            source_path=project_ctx.fs.project_path(),
            project_ctx=project_ctx,
        ),
        GolangWorkspaceBuildPlan(
            name="bh",
            project_ctx=project_ctx,
            prepare_actions=[
                CopyBloodHoundUIAssets(
                    ui_build_name="bh-ui",
                    ce_root_path=project_ctx.fs.project_path(),
                )
            ],
        ),
    ]

    plans.test_plans = [
        YarnTestPlan(
            name="bh-shared-ui",
            source_path=project_ctx.fs.project_path("packages", "javascript", "bh-shared-ui"),
            project_ctx=project_ctx,
            yarn_workspace=True,
        ),
        YarnTestPlan(
            name="js-client-library",
            source_path=project_ctx.fs.project_path("packages", "javascript", "js-client-library"),
            project_ctx=project_ctx,
            yarn_workspace=True,
        ),
        YarnTestPlan(
            name="bh-ui",
            source_path=project_ctx.fs.project_path("cmd", "ui"),
            project_ctx=project_ctx,
            yarn_workspace=True,
        ),
        GolangWorkspaceTestPlan(
            name="bh",
            project_ctx=project_ctx,
        ),
    ]
