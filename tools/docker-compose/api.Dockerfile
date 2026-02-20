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
ARG SHARPHOUND_VERSION=v2.9.0
ARG AZUREHOUND_VERSION=v2.9.1

########
# Package other assets
################
FROM docker.io/library/alpine:3.21 AS hound-builder
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

WORKDIR /tmp/sharphound

# Make some additional directories for minimal container to copy
RUN mkdir -p /opt/bloodhound /etc/bloodhound /var/log
RUN apk --no-cache add p7zip

# Package Sharphound
RUN wget https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip -O sharphound-${SHARPHOUND_VERSION}.zip
RUN wget https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip.sha256 -O sharphound-${SHARPHOUND_VERSION}.zip.sha256

WORKDIR /tmp/azurehound

# Package Azurehound
RUN wget \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_amd64.zip \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_amd64.zip.sha256 \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_arm64.zip \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_arm64.zip.sha256 \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_amd64.zip \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_amd64.zip.sha256 \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_arm64.zip \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_arm64.zip.sha256 \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_amd64.zip \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_amd64.zip.sha256 \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_arm64.zip \
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_arm64.zip.sha256
RUN sha256sum -cw *.sha256
RUN 7z x '*.zip' -oartifacts/*
RUN ls

WORKDIR /tmp/azurehound/artifacts
RUN 7z a -tzip -mx9 azurehound-${AZUREHOUND_VERSION}.zip *
RUN sha256sum azurehound-${AZUREHOUND_VERSION}.zip > azurehound-${AZUREHOUND_VERSION}.zip.sha256

FROM docker.io/library/golang:1.24.13-alpine3.22
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION
ENV GOFLAGS="-buildvcs=false"
WORKDIR /bloodhound
VOLUME [ "/go/pkg/mod" ]

RUN mkdir -p /bhapi/collectors/azurehound /bhapi/collectors/sharphound /bhapi/work
RUN go install github.com/go-delve/delve/cmd/dlv@v1.24.2
RUN go install github.com/air-verse/air@v1.52.3

# api/v2/collectors/[collector-type]/[version] for collector download specifically expects
# '[collector-type]-[version].zip(.sha256)' - all lowercase for embedded files
COPY --from=hound-builder /tmp/sharphound/sharphound-${SHARPHOUND_VERSION}.zip /bhapi/collectors/sharphound/
COPY --from=hound-builder /tmp/sharphound/sharphound-${SHARPHOUND_VERSION}.zip.sha256 /bhapi/collectors/sharphound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-${AZUREHOUND_VERSION}.zip /bhapi/collectors/azurehound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-${AZUREHOUND_VERSION}.zip.sha256 /bhapi/collectors/azurehound/

ENTRYPOINT ["air"]
