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

import { Button, Card } from '@bloodhoundenterprise/doodleui';
import { IconProp } from '@fortawesome/fontawesome-svg-core';
import { faSquarePlus } from '@fortawesome/free-regular-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTagTypeTier, AssetGroupTagTypes } from 'js-client-library';
import { FC } from 'react';
import LargeRightArrow from '../../../components/AppIcon/Icons/LargeRightArrow';
import { ROUTE_TIER_MANAGEMENT_DETAILS } from '../../../routes';
import { useAppNavigate } from '../../../utils';
import { abbreviatedNumber } from '../../../utils/abbreviatedNumber';

type SummaryCardProps = {
    title: string;
    type: AssetGroupTagTypes;
    selectorCount: number | undefined;
    memberCount: number | undefined;
    id: number;
};

const SummaryCard: FC<SummaryCardProps> = ({ title, type, selectorCount, memberCount, id }) => {
    const navigate = useAppNavigate();
    return (
        <Card className='w-full flex px-6 py-4 rounded-xl'>
            <div className='flex-1 flex items-center justify-center truncate min-w-0'>
                <div className='text-2xl font-bold truncate min-w-0'>{title}</div>
            </div>
            <LargeRightArrow className='w-8 h-16' />
            <div className='flex-1 flex flex-col items-center justify-center'>
                <p className='text-lg font-semibold'>Selectors</p>
                <p className='text-2xl font-thin'>{abbreviatedNumber(selectorCount ?? 0)}</p>
            </div>
            <LargeRightArrow className='w-8 h-16' />
            <div className='flex-1 flex flex-col items-center justify-center'>
                <p className='text-lg font-semibold'>Members</p>
                <p className='text-2xl font-thin'>{abbreviatedNumber(memberCount ?? 0)}</p>
            </div>

            <div className='flex-1 flex flex-col items-center justify-center border-l border-black dark:border-white text-sm'>
                <Button
                    variant={'text'}
                    onClick={(e) => {
                        // Prevent event bubbling for the view details action
                        e.stopPropagation();
                        navigate(
                            `/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/${type === AssetGroupTagTypeTier ? 'tier' : 'label'}/${id}`
                        );
                    }}
                    className=' flex items-center space-x-2 hover:underline'>
                    <FontAwesomeIcon icon={faSquarePlus as IconProp} />
                    <p>View Details</p>
                </Button>
            </div>
        </Card>
    );
};

export default SummaryCard;
