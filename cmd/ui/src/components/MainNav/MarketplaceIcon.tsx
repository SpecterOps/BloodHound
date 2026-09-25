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

const MarketplaceIcon = () => (
    <svg
        aria-hidden='true'
        data-testid='marketplace-icon'
        fill='none'
        focusable='false'
        height='24'
        viewBox='0 0 24 24'
        width='24'>
        <rect height='7' rx='1.5' stroke='currentColor' strokeWidth='2' width='7' x='3' y='3' />
        <rect height='7' rx='1.5' stroke='currentColor' strokeWidth='2' width='7' x='14' y='3' />
        <rect height='7' rx='1.5' stroke='currentColor' strokeWidth='2' width='7' x='3' y='14' />
        <rect height='7' rx='1.5' stroke='currentColor' strokeWidth='2' width='7' x='14' y='14' />
    </svg>
);

export default MarketplaceIcon;
