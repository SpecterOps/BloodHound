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
import React from 'react';
import { Link } from 'react-router-dom';
import { AppIcon, CreateMenu } from '../../components';
import { ROUTE_TIER_MANAGEMENT_DETAILS, ROUTE_TIER_MANAGEMENT_SUMMARY } from '../../routes';
import { useAppNavigate } from '../../utils';
import { getTagUrlValue } from './utils';

export const TierActionBar: React.FC<{
    tierId: string;
    labelId: string | undefined;
    selectorId: string | undefined;
    showEditButton?: boolean;
    getSavePath?: (tierId: string | undefined, labelId: string | undefined, selectorId: string | undefined) => string;
}> = ({ tierId, labelId, selectorId, showEditButton, getSavePath }) => {
    const navigate = useAppNavigate();
    const tagId = labelId === undefined ? tierId : labelId;
    return (
        <div className='flex mt-6 gap-8'>
            <div className='flex justify-around basis-2/3'>
                <div className='flex justify-start gap-4 items-center basis-2/3'>
                    <div className='flex items-center align-middle'>
                        <CreateMenu
                            createMenuTitle='View'
                            disabled={!tierId}
                            menuItems={[
                                {
                                    title: 'Summary View',
                                    onClick: () => {
                                        navigate(
                                            `/tier-management/${ROUTE_TIER_MANAGEMENT_SUMMARY}/${getTagUrlValue(labelId)}/${tagId}`
                                        );
                                    },
                                },
                                {
                                    title: 'Detail View',
                                    onClick: () => {
                                        navigate(`/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}`);
                                    },
                                },
                            ]}
                        />
                        <CreateMenu
                            createMenuTitle='Create Selector'
                            disabled={!tierId}
                            menuItems={[
                                {
                                    title: 'Create Selector',
                                    onClick: () => {
                                        navigate(`/tier-management/edit/tag/${tierId}/selector`);
                                    },
                                },
                            ]}
                        />
                        <div className='hidden'>
                            <div>
                                <AppIcon.Info className='mr-4 ml-2' size={24} />
                            </div>
                            <span>
                                To create additional tiers{' '}
                                <Button
                                    variant='text'
                                    asChild
                                    className='p-0 text-base text-secondary dark:text-secondary-variant-2'>
                                    <a href='#'>contact sales</a>
                                </Button>{' '}
                                in order to upgrade for multi-tier analysis.
                            </span>
                        </div>
                    </div>
                </div>
                <div className='flex justify-start basis-1/3'>
                    <input type='text' placeholder='search' className='hidden' />
                </div>
            </div>

            <div className='basis-1/3'>
                {showEditButton && getSavePath && (
                    <Button asChild variant={'secondary'} disabled={showEditButton}>
                        <Link to={getSavePath(tierId, labelId, selectorId)}>Edit</Link>
                    </Button>
                )}
            </div>
        </div>
    );
};
