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
RUN mkdir /.cache && chmod -R go+w /.cache
RUN mkdir /.local && chmod -R go+w /.local
RUN corepack enable
RUN corepack prepare pnpm@9.0.0 --activate

# BloodHound Workspace files
WORKDIR /bloodhound
COPY package.json ./
COPY pnpm-workspace.yaml ./
COPY pnpm-lock.yaml* ./
COPY .npmrc ./
COPY nx.json ./

# Shared Project Files
WORKDIR /bloodhound/packages/javascript
COPY packages/javascript/bh-shared-ui/package.json ./bh-shared-ui/
COPY packages/javascript/bh-shared-ui/project.json ./bh-shared-ui/
COPY packages/javascript/js-client-library/package.json ./js-client-library/
COPY packages/javascript/js-client-library/project.json ./js-client-library/

# BloodHound Project Files
WORKDIR /bloodhound/cmd/ui
COPY cmd/ui/package.json ./
COPY cmd/ui/project.json ./
COPY cmd/ui/vite.config.ts ./
COPY cmd/ui/tsconfig.node.json ./
COPY cmd/ui/tsconfig.json ./
COPY cmd/ui/public ./public
COPY cmd/ui/postcss.config.js ./
COPY cmd/ui/tailwind.config.js ./
COPY cmd/ui/index.html ./

WORKDIR /bloodhound

RUN pnpm install
