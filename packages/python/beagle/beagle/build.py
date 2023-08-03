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

import plans

from beagle.project import ProjectContext


def run_builds(project_ctx: ProjectContext) -> None:
    for build_plan in plans.build_plans:
        if len(project_ctx.runtime.targets) > 0 and build_plan.name not in project_ctx.runtime.targets:
            continue

        try:
            project_ctx.info(f"Executing build plan: {build_plan.name}")

            build_plan.prepare()
            build_plan.build()

            # Output a new line to break output between plans
            project_ctx.info()
        finally:
            build_plan.cleanup()
