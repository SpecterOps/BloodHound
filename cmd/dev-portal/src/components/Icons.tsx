// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

export function StorybookIcon() {
    return (
        <svg viewBox='0 0 32 32' fill='none' aria-hidden='true'>
            <rect x='5' y='5' width='9' height='9' rx='2' />
            <rect x='18' y='5' width='9' height='9' rx='2' />
            <rect x='5' y='18' width='9' height='9' rx='2' />
            <path d='M22.5 19v7M19 22.5h7' />
        </svg>
    );
}

export function DocumentationIcon() {
    return (
        <svg viewBox='0 0 32 32' fill='none' aria-hidden='true'>
            <path d='M7 5.5h12l6 6v15H7v-21Z' />
            <path d='M19 5.5v6h6M11 17h10M11 22h7' />
        </svg>
    );
}
