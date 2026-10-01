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

if [[ $# -ne 2 ]]; then
    echo "usage: $0 <postgres-version> <output-directory>" >&2
    exit 1
fi

postgres_version="$1"
output_directory="$2"
source_archive="postgresql-${postgres_version}.tar.bz2"
source_url="https://ftp.postgresql.org/pub/source/v${postgres_version}/${source_archive}"
work_directory="$(mktemp -d)"

cleanup() {
    rm -rf "${work_directory}"
}
trap cleanup EXIT

curl --fail --location --proto '=https' --remote-name --output-dir "${work_directory}" "${source_url}"
curl --fail --location --proto '=https' --remote-name --output-dir "${work_directory}" "${source_url}.sha256"
(cd "${work_directory}" && shasum -a 256 -c "${source_archive}.sha256")

tar -C "${work_directory}" -xjf "${work_directory}/${source_archive}"
source_directory="${work_directory}/postgresql-${postgres_version}"

rm -rf "${output_directory}"
mkdir -p "${output_directory}"

cd "${source_directory}"
if [[ "$(uname -s)" == "Linux" ]]; then
    export LDFLAGS="${LDFLAGS:-} -Wl,-rpath,'\$\$ORIGIN/../lib'"
fi

./configure \
    --prefix="${output_directory}" \
    --disable-nls \
    --disable-rpath \
    --without-gssapi \
    --without-icu \
    --without-ldap \
    --without-lz4 \
    --without-pam \
    --without-readline \
    --without-zstd
if [[ "$(uname -s)" == "Darwin" ]]; then
    build_jobs="$(sysctl -n hw.ncpu)"
else
    build_jobs="$(nproc)"
fi

make -j"${build_jobs}"
make install
make -C contrib/pg_trgm install
make -C contrib/intarray install

case "$(uname -s)" in
    Darwin)
        if otool -L "${output_directory}/bin/postgres" | grep -qE '/opt/homebrew|/usr/local'; then
            echo "PostgreSQL runtime is not portable; it contains a Homebrew or /usr/local dependency" >&2
            exit 1
        fi
        ;;
    Linux)
        file "${output_directory}/bin/postgres" | grep -q 'ELF 64-bit.*x86-64' || {
            echo "PostgreSQL runtime is not a Linux amd64 executable" >&2
            exit 1
        }
        ;;
esac
