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

import { faWarning } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTag } from 'js-client-library';
import { FC, useContext } from 'react';
import { useHighestPrivilegeTagId, usePZPathParams, useRuleInfo } from '../../../hooks';
import { useAppNavigate } from '../../../utils/searchParams';
import SearchBar from '../Details/SearchBar';
import { PrivilegeZonesContext } from '../PrivilegeZonesContext';
import { RulesAccordion } from './RulesAccordion';

const Details: FC = () => {
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { tagTypeDisplay, tagId: pathTagId, tagDetailsLink, ruleId } = usePZPathParams();
    const tagId = pathTagId ?? topTagId;

    const ruleQuery = useRuleInfo(tagId, ruleId);

    const navigate = useAppNavigate();

    const context = useContext(PrivilegeZonesContext);
    if (!context) {
        throw new Error('Details must be used within a PrivilegeZonesContext.Provider');
    }
    const { InfoHeader, ZoneSelector } = context;

    if (!tagId) return null;

    const handleTagClick = (tag: AssetGroupTag) => navigate(tagDetailsLink(tag.id));

    return (
        <div className='h-full max-h-[80vh]'>
            <div className='flex mt-6'>
                <div className='flex-wrap-reverse basis-2/3 justify-between items-center'>
                    <InfoHeader />
                </div>
            </div>
            <div className='flex gap-8 mt-4 h-full'>
                <div className='flex flex-col gap-2 basis-2/3 bg-neutral-2 py-4 min-w-0 rounded shadow-outer-1 h-full'>
                    <h2 className='font-bold text-xl pl-4 pb-1'>{tagTypeDisplay} Details</h2>
                    <div className='flex justify-between w-full pb-4 border-b border-neutral-3'>
                        <div className='flex gap-6 pl-4'>
                            {ZoneSelector && <ZoneSelector onZoneClick={handleTagClick} />}
                            <div className='flex items-center px-4 rounded h-10 border-contrast border'>
                                TITANCORP.LOCAL
                            </div>
                        </div>
                        <SearchBar showTags={false} />
                    </div>
                    <div className='grow flex overflow-y-scroll overflow-x-hidden max-lg:flex-col'>
                        <div className='basis-1/2'>
                            <RulesAccordion />
                        </div>
                        <div className='basis-1/2'>
                            {ruleQuery.data && ruleQuery.data.disabled_at !== null ? (
                                <div className='flex justify-center items-center gap-2'>
                                    <FontAwesomeIcon icon={faWarning} className='text-orange-500' />
                                    <span>Enable this Rule to see Objects</span>
                                </div>
                            ) : (
                                'Objects Accordion'
                                // <ObjectsAccordion kindCounts={kindCounts} totalCount={777} />
                            )}
                        </div>
                    </div>
                </div>
                <div className='flex basis-1/3 h-full'>{/*Info panes*/}</div>
            </div>
        </div>
    );
};

export default Details;
