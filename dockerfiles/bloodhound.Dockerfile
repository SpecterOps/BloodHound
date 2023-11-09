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
ARG SHARPHOUND_VERSION=v2.0.1
ARG AZUREHOUND_VERSION=v2.1.5

########
# Builder init
################
FROM --platform=$BUILDPLATFORM docker.io/library/node:18-alpine AS deps
ARG version=v999.999.999
ARG checkout_hash=""
ENV PYTHONUNBUFFERED=1
ENV VERSION=${version}
ENV CHECKOUT_HASH=${checkout_hash}
WORKDIR /bloodhound

RUN apk add --update --no-cache python3 git go
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools

COPY . /bloodhound

########
# Build UI
################
FROM deps AS ui-builder

WORKDIR /bloodhound/packages/javascript/bh-shared-ui
RUN yarn install
RUN yarn build

WORKDIR /bloodhound/packages/javascript/js-client-library
RUN yarn install
RUN yarn build

WORKDIR /bloodhound
RUN python3 packages/python/beagle/main.py build bh-ui -v

########
# Build API
################
FROM deps AS api-builder
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0
WORKDIR /bloodhound

COPY --from=ui-builder /bloodhound/dist /bloodhound/dist

RUN python3 packages/python/beagle/main.py build bh -v -d

########
# Package other assets
################
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.16 as hound-builder
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

########
# Package Bloodhound
################
FROM gcr.io/distroless/static-debian11
ARG SHARPHOUND_VERSION
ARG AZUREHOUND_VERSION

COPY dockerfiles/configs/bloodhound.config.json /bloodhound.config.json
COPY --from=api-builder /bloodhound/dist/bhapi /bloodhound
COPY --from=hound-builder /opt/bloodhound /etc/bloodhound /var/log /
COPY --from=hound-builder /tmp/sharphound/sharphound-$SHARPHOUND_VERSION.zip /etc/bloodhound/collectors/sharphound/
COPY --from=hound-builder /tmp/sharphound/sharphound-$SHARPHOUND_VERSION.zip.sha256 /etc/bloodhound/collectors/sharphound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-$AZUREHOUND_VERSION.zip /etc/bloodhound/collectors/azurehound/
COPY --from=hound-builder /tmp/azurehound/artifacts/azurehound-$AZUREHOUND_VERSION.zip.sha256 /etc/bloodhound/collectors/azurehound/

ENTRYPOINT ["/bloodhound", "-configfile", "/bloodhound.config.json"]
