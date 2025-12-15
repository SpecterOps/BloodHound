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

import { FC, useContext } from 'react';
import { useHighestPrivilegeTagId, usePZPathParams } from '../../../hooks';
import { PrivilegeZonesContext } from '../PrivilegeZonesContext';
import SearchBar from './SearchBar';
import { SelectedDetails } from './SelectedDetails';

const Details: FC = () => {
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { zoneId = topTagId?.toString(), tagTypeDisplay, tagId: defaultTagId } = usePZPathParams();
    const tagId = !defaultTagId ? zoneId : defaultTagId;

    const context = useContext(PrivilegeZonesContext);
    if (!context) {
        throw new Error('Details must be used within a PrivilegeZonesContext.Provider');
    }
    const { InfoHeader } = context;

    if (!tagId) return null;

    return (
        <div className='h-full'>
            <div className='flex mt-6'>
                <div className='flex flex-wrap-reverse basis-2/3 justify-between items-center'>
                    <InfoHeader />
                </div>
            </div>
            <div className='flex gap-8 mt-4 h-full'>
                <div className='flex flex-col gap-2 basis-2/3 bg-neutral-2 py-4 min-w-0 rounded shadow-outer-1 h-full'>
                    <h2 className='font-bold text-xl pl-4 pb-1'>{tagTypeDisplay} Details</h2>
                    <div className='flex justify-between w-full pb-4 border-b border-neutral-3'>
                        <div className='flex gap-6 pl-4'>
                            <div className='flex items-center px-4 rounded h-10 border-contrast border'>Tier Zero</div>
                            <div className='flex items-center px-4 rounded h-10 border-contrast border'>
                                TITANCORP.LOCAL
                            </div>
                        </div>
                        <SearchBar showTags={false} />
                    </div>
                    <div className='overflow-hidden'>
                        <div className='w-1/2 max-lg:w-full'>
                            <ul className='h-dvh overflow-y-scroll'></ul>
                        </div>
                        <div className='w-1/2 max-lg:w-full'>
                            <ul className='h-dvh overflow-y-scroll'></ul>
                        </div>
                    </div>
                </div>
                <div className='flex basis-1/3 h-full'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Details;
