#!/usr/bin/env bash
# Copyright 2026 Specter Ops, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
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

set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "usage: $0 <self-contained-postgres-runtime-root>" >&2
    exit 1
fi

runtime_root="$1"
repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
archive_path="${repository_root}/cmd/api/src/standalone/assets/postgres-runtime.tar.gz"

for required_path in bin/createdb bin/initdb bin/pg_ctl lib share; do
    if [[ ! -e "${runtime_root}/${required_path}" ]]; then
        echo "PostgreSQL runtime is missing ${required_path}: ${runtime_root}" >&2
        exit 1
    fi
done

if otool -L "${runtime_root}/bin/postgres" | grep -qE '/opt/homebrew|/usr/local'; then
    echo "PostgreSQL runtime is not portable; build it with scripts/standalone/build-postgres-runtime.sh" >&2
    exit 1
fi

mkdir -p "$(dirname "${archive_path}")"
rm -f "${archive_path}"
# Materialize links so the Go extractor can reject link entries safely while
# still receiving every dylib required by PostgreSQL on macOS.
tar -h -C "${runtime_root}" -czf "${archive_path}" bin lib share
shasum -a 256 "${archive_path}"
