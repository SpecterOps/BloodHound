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
import capitalize from 'lodash/capitalize';
import { FC } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import { AppLink } from '../../../components/Navigation';
import { useHighestPrivilegeTagId, useOwnedTag } from '../../../hooks';
import SelectorForm from './SelectorForm';
import TagForm from './TagForm';

const Save: FC = () => {
    const location = useLocation();
    const { tierId, labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const showSelectorForm = location.pathname.includes('selector');
    const tagValue = location.pathname.includes('label') ? 'label' : 'tier';
    const capitalizedTagValue = capitalize(tagValue);
    const captitalizedPluralTagValue = capitalizedTagValue + 's';
    const topTagId = useHighestPrivilegeTagId();
    const ownedId = useOwnedTag()?.id;
    return (
        <div>
            <Breadcrumb className='my-6'>
                <BreadcrumbList>
                    <BreadcrumbItem>
                        <BreadcrumbLink asChild>
                            <AppLink
                                to={`/zone-management/details/${tagValue}/${tagValue === 'tier' ? topTagId : ownedId}`}>
                                {captitalizedPluralTagValue}
                            </AppLink>
                        </BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator />
                    {showSelectorForm ? (
                        <>
                            <BreadcrumbItem>
                                <BreadcrumbLink asChild>
                                    <AppLink to={`/zone-management/save/${tagValue}/${tagId}`}>
                                        {`${capitalizedTagValue} Details`}
                                    </AppLink>
                                </BreadcrumbLink>
                            </BreadcrumbItem>
                            <BreadcrumbSeparator />
                            <BreadcrumbItem>
                                <BreadcrumbPage>Selector</BreadcrumbPage>
                            </BreadcrumbItem>
                        </>
                    ) : (
                        <>
                            <BreadcrumbItem>
                                <BreadcrumbPage>{`${capitalizedTagValue} Details`}</BreadcrumbPage>
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
