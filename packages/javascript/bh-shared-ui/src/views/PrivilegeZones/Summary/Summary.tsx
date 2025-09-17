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

import { Button } from '@bloodhoundenterprise/doodleui';
import { FC, useContext } from 'react';
import { UseQueryResult } from 'react-query';
import { AppLink } from '../../../components';
import { usePZPathParams, useTagsQuery } from '../../../hooks';
import { privilegeZonesPath, summaryPath } from '../../../routes';
import { useAppNavigate } from '../../../utils';
import { getSavePath } from '../Details/Details';
import { SelectedDetails } from '../Details/SelectedDetails';
import { PrivilegeZonesContext } from '../PrivilegeZonesContext';
import SummaryList from './SummaryList';

export const getEditButtonState = (memberId?: string, selectorsQuery?: UseQueryResult, tagsQuery?: UseQueryResult) => {
    return (
        !!memberId ||
        (selectorsQuery?.isLoading && tagsQuery?.isLoading) ||
        (selectorsQuery?.isError && tagsQuery?.isError)
    );
};

const Summary: FC = () => {
    const navigate = useAppNavigate();
    const { zoneId, labelId, selectorId, tagId, tagType, tagTypeDisplayPlural } = usePZPathParams();
    const tagsQuery = useTagsQuery();
    const saveLink = getSavePath(zoneId, labelId, selectorId);
    const context = useContext(PrivilegeZonesContext);
    if (!context) {
        throw new Error('Details must be used within a PrivilegeZonesContext.Provider');
    }
    const { InfoHeader } = context;

    return (
        <div className='h-full'>
            <div className='flex mt-6 gap-8'>
                <InfoHeader />
                <div className='basis-1/3'>
                    {saveLink ? (
                        <Button asChild variant={'secondary'}>
                            <AppLink data-testid='privilege-zones_edit-button' to={saveLink}>
                                Edit
                            </AppLink>
                        </Button>
                    ) : (
                        <Button variant={'secondary'} disabled>
                            Edit
                        </Button>
                    )}
                </div>
            </div>
            <div className='flex gap-8 mt-4 w-full h-full'>
                <div className='basis-2/3'>
                    <SummaryList
                        title={tagTypeDisplayPlural}
                        listQuery={tagsQuery}
                        selected={tagId}
                        onSelect={(tagId) => navigate(`/${privilegeZonesPath}/${tagType}/${tagId}/${summaryPath}`)}
                    />
                </div>
                <div className='basis-1/3'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Summary;
