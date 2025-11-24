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

import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { AppLink } from '../../../components/Navigation';
import { useHighestPrivilegeTagId, useOwnedTagId, usePZPathParams } from '../../../hooks';
import { rulesPath } from '../../../routes';
import SelectorForm from './SelectorForm';
import TagForm from './TagForm';

const Save: FC = () => {
    const showSelectorForm = location.pathname.includes(rulesPath);
    const { tagTypeDisplay, tagTypeDisplayPlural, tagId, tagDetailsLink, tagEditLink, isZonePage } = usePZPathParams();

    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const ownedId = useOwnedTagId();

    return (
        <div>
            <Breadcrumb className='my-6'>
                <BreadcrumbList>
                    <BreadcrumbItem>
                        <BreadcrumbLink asChild>
                            {!topTagId || !ownedId ? (
                                <span data-testid='privilege-zones_save_details-breadcrumb'>
                                    {tagTypeDisplayPlural}
                                </span>
                            ) : (
                                <AppLink
                                    data-testid='privilege-zones_save_details-breadcrumb'
                                    to={tagDetailsLink(isZonePage ? topTagId : ownedId)}>
                                    {tagTypeDisplayPlural}
                                </AppLink>
                            )}
                        </BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator />
                    {showSelectorForm ? (
                        <>
                            <BreadcrumbItem>
                                <BreadcrumbLink asChild>
                                    <AppLink data-testid='privilege-zones_save_tag-breadcrumb' to={tagEditLink(tagId)}>
                                        {`${tagTypeDisplay} Details`}
                                    </AppLink>
                                </BreadcrumbLink>
                            </BreadcrumbItem>
                            <BreadcrumbSeparator />
                            <BreadcrumbItem>
                                <BreadcrumbPage>Rule</BreadcrumbPage>
                            </BreadcrumbItem>
                        </>
                    ) : (
                        <>
                            <BreadcrumbItem>
                                <BreadcrumbPage>{`${tagTypeDisplay} Details`}</BreadcrumbPage>
                            </BreadcrumbItem>
                        </>
                    )}
                </BreadcrumbList>
            </Breadcrumb>
            {showSelectorForm ? <SelectorForm /> : <TagForm />}
        </div>
    );
};

export default Save;
