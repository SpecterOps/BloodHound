#!/usr/bin/env bash

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

DLV_FLAGS="--listen=:2345 --headless=true --log=true --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --accept-multiclient --api-version=2"
go build -ldflags="-extldflags '-static'" -gcflags "all=-N -l" -o /bhapi github.com/specterops/bloodhound/src/cmd/bhapi
/go/bin/dlv $DLV_FLAGS exec /bhapi -- -configfile local-harnesses/build.config.json
