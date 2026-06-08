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
#Manually set/supply build arg versions if you want to bundle older collectors
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

FROM docker.io/library/alpine:3.21 AS sharphound-latest-release
ADD https://api.github.com/repos/SpecterOps/SharpHound/releases/latest /tmp/sharphound-release.json
RUN set -eux; \
    sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /tmp/sharphound-release.json > /tmp/version

FROM docker.io/library/alpine:3.21 AS azurehound-latest-release
ADD https://api.github.com/repos/SpecterOps/AzureHound/releases/latest /tmp/azurehound-release.json
RUN set -eux; \
    sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /tmp/azurehound-release.json > /tmp/version

########
# Validate collector versions
################
FROM docker.io/library/alpine:3.21 AS collector-version-validator
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

COPY --from=sharphound-latest-release /tmp/version /tmp/sharphound-latest-version
COPY --from=azurehound-latest-release /tmp/version /tmp/azurehound-latest-version

# Build arg versions take precedence over latest version
RUN set -eux; \
    sharphound_version="${SHARPHOUND_VERSION:-$(cat /tmp/sharphound-latest-version)}"; \
    azurehound_version="${AZUREHOUND_VERSION:-$(cat /tmp/azurehound-latest-version)}"; \
    echo "${sharphound_version}" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$' || { echo "SHARPHOUND_VERSION must match vX.Y.Z" >&2; exit 1; }; \
    echo "${azurehound_version}" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$' || { echo "AZUREHOUND_VERSION must match vX.Y.Z" >&2; exit 1; }; \
    printf 'SHARPHOUND_VERSION=%s\nAZUREHOUND_VERSION=%s\n' "${sharphound_version}" "${azurehound_version}" > /tmp/collector-versions

########
# Package remote assets
################
FROM collector-version-validator AS hound-builder
RUN apk --no-cache add p7zip

# Make some additional directories for minimal container to copy
RUN mkdir -p /opt/bloodhound /etc/bloodhound /var/log

# Download SharpHound artifacts
WORKDIR /tmp/sharphound
RUN set -eux; \
    . /tmp/collector-versions; \
    wget "https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip"; \
    wget "https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip.sha256"; \
    sha256sum -cw *.sha256

# Package SharpHound in /tmp/sharphound/dist
WORKDIR /tmp/sharphound
RUN set -eux; \
    . /tmp/collector-versions; \
    mkdir -p dist; \
    mv "SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip" "dist/sharphound-${SHARPHOUND_VERSION}.zip"; \
    (cd dist && sha256sum "sharphound-${SHARPHOUND_VERSION}.zip" > "sharphound-${SHARPHOUND_VERSION}.zip.sha256")

# Download AzureHound artifacts
WORKDIR /tmp/azurehound
RUN set -eux; \
    . /tmp/collector-versions; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_amd64.zip"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_amd64.zip.sha256"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_arm64.zip"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_arm64.zip.sha256"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_amd64.zip"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_amd64.zip.sha256"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_arm64.zip"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_linux_arm64.zip.sha256"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_amd64.zip"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_amd64.zip.sha256"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_arm64.zip"; \
    wget "https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_arm64.zip.sha256"; \
    sha256sum -cw *.sha256

# Package AzureHound in /tmp/azurehound/dist
RUN set -eux; \
    . /tmp/collector-versions; \
    mkdir -p artifacts dist; \
    7z x '*.zip' -oartifacts/*; \
    (cd artifacts && 7z a -tzip -mx9 "../dist/azurehound-${AZUREHOUND_VERSION}.zip" *); \
    (cd dist && sha256sum "azurehound-${AZUREHOUND_VERSION}.zip" > "azurehound-${AZUREHOUND_VERSION}.zip.sha256")

FROM docker.io/library/golang:1.26.4-alpine3.22
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION
ENV GOFLAGS="-buildvcs=false"
WORKDIR /bloodhound
VOLUME [ "/go/pkg/mod" ]

RUN mkdir -p /bhapi/collectors/azurehound /bhapi/collectors/sharphound /bhapi/work
RUN go install github.com/go-delve/delve/cmd/dlv@v1.26.1
RUN go install github.com/air-verse/air@v1.52.3

# api/v2/collectors/[collector-type]/[version] for collector download specifically expects
# '[collector-type]-[version].zip(.sha256)' - all lowercase for embedded files
COPY --from=hound-builder /tmp/sharphound/dist/ /bhapi/collectors/sharphound/
COPY --from=hound-builder /tmp/azurehound/dist/ /bhapi/collectors/azurehound/

ENTRYPOINT ["air"]
