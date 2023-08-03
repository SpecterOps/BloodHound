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

########
# Global build args
################
ARG SHARPHOUND_VERSION=v2.0.0
ARG AZUREHOUND_VERSION=v2.0.4


########
# Package other assets
################
FROM alpine:3.16 as hound-builder
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

WORKDIR /tmp/sharphound

# Make some additional directories for minimal container to copy
RUN mkdir -p /opt/bloodhound /etc/bloodhound /var/log
RUN apk --no-cache add p7zip

# Package Sharphound
RUN wget https://github.com/BloodHoundAD/SharpHound/releases/download/$SHARPHOUND_VERSION/SharpHound-$SHARPHOUND_VERSION.zip -O sharphound-$SHARPHOUND_VERSION.zip
RUN sha256sum sharphound-$SHARPHOUND_VERSION.zip > sharphound-$SHARPHOUND_VERSION.zip.sha256

WORKDIR /tmp/azurehound

# Package Azurehound
RUN wget \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-darwin-amd64.zip \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-darwin-amd64.zip.sha256 \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-darwin-arm64.zip \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-darwin-arm64.zip.sha256 \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-linux-amd64.zip \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-linux-amd64.zip.sha256 \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-linux-arm64.zip \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-linux-arm64.zip.sha256 \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-windows-amd64.zip \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-windows-amd64.zip.sha256 \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-windows-arm64.zip \
  https://github.com/BloodHoundAD/AzureHound/releases/download/$AZUREHOUND_VERSION/azurehound-windows-arm64.zip.sha256
RUN sha256sum -cw *.sha256
RUN 7z x '*.zip' -oartifacts/*
RUN ls

WORKDIR /tmp/azurehound/artifacts
RUN 7z a -tzip -mx9 azurehound-$AZUREHOUND_VERSION.zip azurehound-*
RUN sha256sum azurehound-$AZUREHOUND_VERSION.zip > azurehound-$AZUREHOUND_VERSION.zip.sha256

FROM golang:1.20
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION
WORKDIR /bloodhound
VOLUME [ "/go/pkg/mod" ]
RUN mkdir -p /bhapi/collectors/azurehound /bhapi/collectors/sharphound /bhapi/work
RUN go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
  && apt-get -y install --no-install-recommends entr
# Fix for running git ls-files in this container
RUN git config --global --add safe.directory /bloodhound
COPY scripts/file-watcher.sh scripts/api-rebuild.sh scripts/delve-rebuild.sh /
COPY --from=hound-builder /tmp/sharphound/sharphound-$SHARPHOUND_VERSION.zip /bhapi/collectors/sharphound/
COPY --from=hound-builder /tmp/sharphound/sharphound-$SHARPHOUND_VERSION.zip.sha256 /bhapi/collectors/sharphound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-$AZUREHOUND_VERSION.zip /bhapi/collectors/azurehound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-$AZUREHOUND_VERSION.zip.sha256 /bhapi/collectors/azurehound/
