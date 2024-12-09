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

FROM docker.io/library/node:22-alpine AS base

# Setup
RUN mkdir /.yarn && \
  chmod -R go+w /.yarn \
  mkdir /.cache && \
  chmod -R go+w /.cache && \
  correpack enable && \
  corepack prepare yarn@stable --activate

# BloodHound Workspace files
WORKDIR /bloodhound

COPY package.json \
  yarn.lock \
  .yarnrc.yml \
  .yarn \
  ./

# Shared Project Files
WORKDIR /bloodhound/packages/javascript
COPY packages/javascript/bh-shared-ui/package.json ./bh-shared-ui/
COPY packages/javascript/js-client-library/package.json ./js-client-library/

# BloodHound Project Files
WORKDIR /bloodhound/cmd/ui
COPY cmd/ui/package.json \
  cmd/ui/vite.config.ts \
  cmd/ui/tsconfig.node.json \
  cmd/ui/tsconfig.json \
  cmd/ui/index.html \
  ./
COPY cmd/ui/public ./public

RUN yarn
