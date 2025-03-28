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
ARG SHARPHOUND_VERSION=v2.6.1
ARG AZUREHOUND_VERSION=v2.3.0

########
# Golang Image
################
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.23-alpine3.20 AS godeps

########
# Builder init
################
FROM --platform=$BUILDPLATFORM docker.io/library/node:22-alpine3.20 AS deps
ARG version=v999.999.999
ARG checkout_hash=""
ENV SB_LOG_LEVEL=debug
ENV SB_VERSION=${version}
ENV CHECKOUT_HASH=${checkout_hash}
WORKDIR /bloodhound

RUN apk add --update --no-cache git

COPY --from=godeps /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

COPY . /bloodhound
RUN go run github.com/specterops/bloodhound/packages/go/stbernard deps

########
# Build
################
FROM deps AS builder
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
ENV SB_VERSION=${version}
WORKDIR /bloodhound

RUN go run github.com/specterops/bloodhound/packages/go/stbernard build --os ${TARGETOS} --arch ${TARGETARCH}

########
# Package other assets
################
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.20 AS hound-builder
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

WORKDIR /tmp/sharphound

# Make some additional directories for minimal container to copy
RUN mkdir -p /opt/bloodhound /etc/bloodhound /var/log
RUN apk --no-cache add p7zip

# Package Sharphound
RUN wget https://github.com/SpecterOps/SharpHound/releases/download/$SHARPHOUND_VERSION/SharpHound_$SHARPHOUND_VERSION_windows_x86.zip -O sharphound-$SHARPHOUND_VERSION.zip
RUN wget https://github.com/SpecterOps/SharpHound/releases/download/$SHARPHOUND_VERSION/SharpHound_$SHARPHOUND_VERSION_windows_x86.zip.sha256 -O sharphound-$SHARPHOUND_VERSION.zip.sha256

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
RUN 7z a -tzip -mx9 azurehound-$AZUREHOUND_VERSION.zip AzureHound_*
RUN sha256sum azurehound-$AZUREHOUND_VERSION.zip > azurehound-$AZUREHOUND_VERSION.zip.sha256

########
# Package Bloodhound
################
FROM gcr.io/distroless/static-debian11 AS bloodhound
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

COPY dockerfiles/configs/bloodhound.config.json /bloodhound.config.json
COPY --from=builder /bloodhound/dist/bhapi /bloodhound
COPY --from=hound-builder /opt/bloodhound /etc/bloodhound /var/log /
COPY --from=hound-builder /tmp/sharphound/sharphound-$SHARPHOUND_VERSION.zip /etc/bloodhound/collectors/sharphound/
COPY --from=hound-builder /tmp/sharphound/sharphound-$SHARPHOUND_VERSION.zip.sha256 /etc/bloodhound/collectors/sharphound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-$AZUREHOUND_VERSION.zip /etc/bloodhound/collectors/azurehound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-$AZUREHOUND_VERSION.zip.sha256 /etc/bloodhound/collectors/azurehound/

ENTRYPOINT ["/bloodhound", "-configfile", "/bloodhound.config.json"]
