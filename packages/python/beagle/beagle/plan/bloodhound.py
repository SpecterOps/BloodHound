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
import shutil

from beagle.facts import GIT_KEEP_FILES
from beagle.plan import Action
from beagle.project import ProjectContext


class CopyBloodHoundUIAssets(Action):
    def __init__(self, ui_build_name: str, ce_root_path: str):
        self.ui_build_name = ui_build_name
        self.ce_root_path = ce_root_path

    def run(self, plan_name: str, project_ctx: ProjectContext) -> None:
        built_ui_assets_path = project_ctx.fs.artifact_path(self.ui_build_name)

        if not os.path.lexists(built_ui_assets_path):
            raise Exception(f"No UI assets found at {built_ui_assets_path}")

        ui_assets_path = os.path.join(self.ce_root_path, "cmd", "api", "src", "api", "static", "assets")
        project_ctx.info(f"Copying UI assets from {built_ui_assets_path} to {ui_assets_path}")

        for filename in os.listdir(ui_assets_path):
            file_path = os.path.join(ui_assets_path, filename)

            # Ignore git keep files
            if filename in GIT_KEEP_FILES:
                continue

            if os.path.isdir(file_path):
                shutil.rmtree(os.path.join(ui_assets_path, filename))
            else:
                os.remove(file_path)

        for filename in os.listdir(built_ui_assets_path):
            file_path = os.path.join(built_ui_assets_path, filename)
            file_target_path = os.path.join(ui_assets_path, filename)

            if os.path.isdir(file_path):
                shutil.copytree(file_path, file_target_path)
            else:
                shutil.copy(file_path, file_target_path)
