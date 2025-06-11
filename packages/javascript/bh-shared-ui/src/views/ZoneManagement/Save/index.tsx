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
import { Link, useLocation, useParams } from 'react-router-dom';
import { getTagUrlValue } from '../../../utils/tagUrlValue';
import SelectorForm from './SelectorForm';
import { TagForm } from './TagForm';

const SaveView: FC = () => {
    const location = useLocation();
    const { tierId, labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const showSelectorForm = location.pathname.includes('selector');

    return (
        <div>
            {/* TODO: REMOVE HIDDEN CLASS WHEN TAG FORM IS FUNCTIONAL */}
            <Breadcrumb className='mb-2 hidden'>
                <BreadcrumbList>
                    <BreadcrumbItem>
                        <BreadcrumbLink asChild>
                            <Link to={`/zone-management/details`}>Tiers</Link>
                        </BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator />
                    {showSelectorForm ? (
                        <>
                            <BreadcrumbItem>
                                <BreadcrumbLink asChild>
                                    <Link to={`/zone-management/save/${getTagUrlValue(labelId)}/${tagId}`}>
                                        Tier Details
                                    </Link>
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
                                <BreadcrumbPage>Tier Details</BreadcrumbPage>
                            </BreadcrumbItem>
                        </>
                    )}
                </BreadcrumbList>
            </Breadcrumb>
            {showSelectorForm ? <SelectorForm /> : <TagForm />}
        </div>
    );
};

export default SaveView;
