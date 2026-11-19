# syntax=docker/dockerfile:1-labs

# Copyright 2025 Specter Ops, Inc.
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
# Optional collector versions; empty values use the latest stable GitHub release.
ARG SHARPHOUND_VERSION=""
ARG AZUREHOUND_VERSION=""
# Create an empty metadata stage for explicit versions
ARG SHARPHOUND_RELEASE_SOURCE=${SHARPHOUND_VERSION:+explicit-release}
ARG AZUREHOUND_RELEASE_SOURCE=${AZUREHOUND_VERSION:+explicit-release}

FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.21 AS explicit-release
RUN touch /tmp/version

FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.21 AS sharphound-latest-release
ADD https://api.github.com/repos/SpecterOps/SharpHound/releases/latest /tmp/sharphound-release.json
RUN set -eux; \
    sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /tmp/sharphound-release.json > /tmp/version

FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.21 AS azurehound-latest-release
ADD https://api.github.com/repos/SpecterOps/AzureHound/releases/latest /tmp/azurehound-release.json
RUN set -eux; \
    sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /tmp/azurehound-release.json > /tmp/version

# Select explicit or latest release stage, buildkit will prune the unused stages, avoiding extra api calls.
FROM --platform=$BUILDPLATFORM ${SHARPHOUND_RELEASE_SOURCE:-sharphound-latest-release} AS sharphound-release
FROM --platform=$BUILDPLATFORM ${AZUREHOUND_RELEASE_SOURCE:-azurehound-latest-release} AS azurehound-release

########
# Resolve and validate collector versions
################
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.21 AS collector-version-resolver
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

COPY --from=sharphound-release /tmp/version /tmp/sharphound-latest-version
COPY --from=azurehound-release /tmp/version /tmp/azurehound-latest-version

# Build arg versions take precedence over latest version
RUN set -eu; \
    sharphound_version="${SHARPHOUND_VERSION:-$(cat /tmp/sharphound-latest-version)}"; \
    azurehound_version="${AZUREHOUND_VERSION:-$(cat /tmp/azurehound-latest-version)}"; \
    sharphound_version="$(printf '%s' "${sharphound_version}" | tr -d '[:space:]')"; \
    azurehound_version="$(printf '%s' "${azurehound_version}" | tr -d '[:space:]')"; \
    version_regex='^v[0-9]+\.[0-9]+\.[0-9]+$'; \
    if ! printf '%s\n' "${sharphound_version}" | grep -Eq "$version_regex"; then \
        echo "SHARPHOUND_VERSION must match vX.Y.Z" >&2; exit 1; \
    fi; \
    if ! printf '%s\n' "${azurehound_version}" | grep -Eq "$version_regex"; then \
        echo "AZUREHOUND_VERSION must match vX.Y.Z" >&2; exit 1; \
    fi; \
    printf 'SHARPHOUND_VERSION=%s\nAZUREHOUND_VERSION=%s\n' "${sharphound_version}" "${azurehound_version}" > /collector-versions; \
    echo "Resolved SharpHound version: ${sharphound_version}"; \
    echo "Resolved AzureHound version: ${azurehound_version}"

########
# Package remote assets
################
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.21 AS hound-builder
RUN apk --no-cache add p7zip

COPY --from=collector-version-resolver /collector-versions /collector-versions

# Download SharpHound artifacts
WORKDIR /tmp/sharphound
RUN set -eux; \
    . /collector-versions; \
    wget "https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip"; \
    wget "https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip.sha256"; \
    sha256sum -cw *.sha256

# Package SharpHound in /tmp/sharphound/dist
RUN set -eux; \
    . /collector-versions; \
    mkdir -p dist; \
    mv "SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip" "dist/sharphound-${SHARPHOUND_VERSION}.zip"; \
    (cd dist && sha256sum "sharphound-${SHARPHOUND_VERSION}.zip" > "sharphound-${SHARPHOUND_VERSION}.zip.sha256")

# Download Azurehound artifacts
WORKDIR /tmp/azurehound
RUN set -eux; \
    . /collector-versions; \
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
    . /collector-versions; \
    mkdir -p artifacts dist; \
    7z x '*.zip' -oartifacts/*; \
    (cd artifacts && 7z a -tzip -mx9 "../dist/azurehound-${AZUREHOUND_VERSION}.zip" *); \
    (cd dist && sha256sum "azurehound-${AZUREHOUND_VERSION}.zip" > "azurehound-${AZUREHOUND_VERSION}.zip.sha256")

########
# UI Build
################
FROM --platform=$BUILDPLATFORM docker.io/library/node:22-alpine3.20 AS ui-builder

WORKDIR /build
COPY --parents constraints.pro package.json **/package.json yarn* .yarn*  ./
RUN yarn install

COPY --parents cmd/ui packages/javascript ./
RUN yarn build

########
# Version Build
################
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.4-alpine3.24 AS ldflag-builder
ENV VERSION_PKG="github.com/specterops/bloodhound/cmd/api/src/version"
RUN apk add --update --no-cache git
WORKDIR /build
COPY .git ./.git

# sort by semver version to grab latest and convert to required ldflags
# (see https://git-scm.com/docs/git-config#Documentation/git-config.txt-versionsortsuffix)
RUN git --no-pager -c 'versionsort.suffix=-rc' tag --list v*.*.* --sort=-v:refname | head -n 1 | sed 's/^v//' | awk \
    -F'[.+-]' \
    -v pkg="$VERSION_PKG" \
    '{ major = $1; minor = $2; patch = $3; pre = ""; if ($4) pre = $4; \
    printf("-X '\''%s.majorVersion=%s'\'' ", pkg, major); \
    printf("-X '\''%s.minorVersion=%s'\'' ", pkg, minor); \
    printf("-X '\''%s.patchVersion=%s'\''", pkg, patch); \
    if (pre != "") \
    printf(" -X '\''%s.prereleaseVersion=%s'\''", pkg, pre); \
    }' > LDFLAGS

########
# API Build
################
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.4-alpine3.24 AS api-builder

ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

RUN apk add --update --no-cache git
RUN mkdir -p /opt/bloodhound /etc/bloodhound /var/log

WORKDIR /build
COPY --parents go* cmd/api packages/go server ./
COPY --from=ldflag-builder /build/LDFLAGS ./
COPY --from=ui-builder /build/cmd/ui/dist ./cmd/api/src/api/static/assets
RUN --mount=type=cache,target=/go/pkg/mod go build -C cmd/api/src -o /bloodhound -ldflags "$(cat LDFLAGS)" github.com/specterops/bloodhound/cmd/api/src/cmd/bhapi

########
# Package BloodHound
################
FROM gcr.io/distroless/static-debian12 AS bloodhound

COPY --from=api-builder /bloodhound /opt/bloodhound /etc/bloodhound /var/log /
COPY dockerfiles/configs/bloodhound.config.json /bloodhound.config.json

# api/v2/collectors/[collector-type]/[version] for collector download specifically expects
# '[collector-type]-[version].zip(.sha256)' - all lowercase for embedded files
COPY --from=hound-builder /tmp/sharphound/dist/ /etc/bloodhound/collectors/sharphound/
COPY --from=hound-builder /tmp/azurehound/dist/ /etc/bloodhound/collectors/azurehound/

ENTRYPOINT ["/bloodhound", "-configfile", "/bloodhound.config.json"]
