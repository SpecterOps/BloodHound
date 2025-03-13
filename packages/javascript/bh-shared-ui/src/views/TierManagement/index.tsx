// Copyright 2025 Specter Ops, Inc.
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
import Details from './Details';

export const TierManagement: FC = () => {
    return (
        <div className='min-h-full min-w-full px-8'>
            <h1 className='text-4xl font-bold pt-8'>Tier Management</h1>
            <p className='mt-6'>
                <span>Define and manage selectors to dynamically gather objects based on criteria.</span>
                <br />
                <span>Ensure selectors capture the right assets for groups assignments or review.</span>
            </p>

            <div className='flex flex-col'>
                <div className='flex gap-4 mt-6'>
                    <div className='text-lg underline'>Tiers</div>
                    <div className='text-lg'>Labels</div>
                    <div className='text-lg'>Certifications</div>
                    <div className='text-lg'>History</div>
                </div>
                <Details />
            </div>
        </div>
    );
};

export * from './Create';
export * from './Edit';
