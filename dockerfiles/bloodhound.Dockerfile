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
ARG SHARPHOUND_VERSION=v2.9.0
ARG AZUREHOUND_VERSION=v2.9.0

########
# Package remote assets
################
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.20 AS hound-builder
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

RUN apk --no-cache add p7zip
RUN mkdir -p /tmp/sharphound /tmp/azurehound

ADD https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip /tmp/sharphound/sharphound-${SHARPHOUND_VERSION}.zip
ADD https://github.com/SpecterOps/SharpHound/releases/download/${SHARPHOUND_VERSION}/SharpHound_${SHARPHOUND_VERSION}_windows_x86.zip.sha256 /tmp/sharphound/sharphound-${SHARPHOUND_VERSION}.zip.sha256

ADD https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_darwin_amd64.zip \
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
  https://github.com/SpecterOps/AzureHound/releases/download/${AZUREHOUND_VERSION}/AzureHound_${AZUREHOUND_VERSION}_windows_arm64.zip.sha256 \
  /tmp/azurehound/

WORKDIR /tmp/azurehound
RUN sha256sum -cw *.sha256
RUN 7z x '*.zip' -oartifacts/*

WORKDIR /tmp/azurehound/artifacts
RUN 7z a -tzip -mx9 azurehound-${AZUREHOUND_VERSION}.zip *
RUN sha256sum azurehound-${AZUREHOUND_VERSION}.zip > azurehound-${AZUREHOUND_VERSION}.zip.sha256

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
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.24.12-alpine3.22 AS ldflag-builder
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
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.24.12-alpine3.22 AS api-builder

ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

RUN apk add --update --no-cache git
RUN mkdir -p /opt/bloodhound /etc/bloodhound /var/log

WORKDIR /build
COPY --parents go* cmd/api packages/go ./
COPY --from=ldflag-builder /build/LDFLAGS ./
COPY --from=ui-builder /build/cmd/ui/dist ./cmd/api/src/api/static/assets
RUN --mount=type=cache,target=/go/pkg/mod go build -C cmd/api/src -o /bloodhound -ldflags "$(cat LDFLAGS)" github.com/specterops/bloodhound/cmd/api/src/cmd/bhapi

########
# Package BloodHound
################
FROM gcr.io/distroless/static-debian11 AS bloodhound
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

COPY --from=api-builder /bloodhound /opt/bloodhound /etc/bloodhound /var/log /
COPY dockerfiles/configs/bloodhound.config.json /bloodhound.config.json

# api/v2/collectors/[collector-type]/[version] for collector download specifically expects
# '[collector-type]-[version].zip(.sha256)' - all lowercase for embedded files
COPY --from=hound-builder /tmp/sharphound/sharphound-${SHARPHOUND_VERSION}.zip /etc/bloodhound/collectors/sharphound/
COPY --from=hound-builder /tmp/sharphound/sharphound-${SHARPHOUND_VERSION}.zip.sha256 /etc/bloodhound/collectors/sharphound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-${AZUREHOUND_VERSION}.zip /etc/bloodhound/collectors/azurehound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-${AZUREHOUND_VERSION}.zip.sha256 /etc/bloodhound/collectors/azurehound/

ENTRYPOINT ["/bloodhound", "-configfile", "/bloodhound.config.json"]
