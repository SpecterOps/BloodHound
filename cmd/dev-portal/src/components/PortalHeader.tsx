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

import { Link } from 'react-router-dom';

export function PortalHeader() {
    return (
        <header className='portal-header'>
            <div className='brand'>
                <span className='brand-mark' aria-hidden='true'>
                    BH
                </span>
                <span>BloodHound</span>
                <span className='brand-divider' aria-hidden='true' />
                <Link className='portal-header-home' to='/'>
                    Developer Portal
                </Link>
            </div>
            <span className='header-label'>Internal resources</span>
        </header>
    );
}
