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

import { FC } from 'react';

const References: FC = () => {
    return (
        <div className='overflow-x-auto'>
            <a
                target='_blank'
                rel='noopener noreferrer'
                href='https://blog.netspi.com/attacking-azure-with-custom-script-extensions/'>
                https://blog.netspi.com/attacking-azure-with-custom-script-extensions/
            </a>
            <br />
            <a
                target='_blank'
                rel='noopener noreferrer'
                href='https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#owner'>
                https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#owner
            </a>
        </div>
    );
};

export default References;
