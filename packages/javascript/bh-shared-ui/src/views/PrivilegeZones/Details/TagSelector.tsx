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

import { AssetGroupTag } from 'js-client-library';
import { FC } from 'react';
import { usePZPathParams } from '../../../hooks/usePZParams/usePZPathParams';
import { useAppNavigate } from '../../../utils/searchParams';
import { usePZContext } from '../PrivilegeZonesContext';
import { TagTabValue } from '../utils';
import { useSelectedDetailsTabsContext } from './SelectedDetailsTabs/SelectedDetailsTabsContext';

const TagSelector: FC = () => {
    const navigate = useAppNavigate();

    const { LabelSelector, ZoneSelector } = usePZContext();
    const { setSelectedDetailsTab } = useSelectedDetailsTabsContext();

    const { isZonePage, isLabelPage, tagDetailsLink } = usePZPathParams();

    const handleTagClick = (tag: AssetGroupTag) => {
        setSelectedDetailsTab(TagTabValue);
        navigate(tagDetailsLink(tag.id));
    };

    if (isZonePage && ZoneSelector) return <ZoneSelector onZoneClick={handleTagClick} />;
    if (isZonePage && !ZoneSelector) return <span className='uppercase font-medium'>Tier Zero</span>;

    if (isLabelPage && LabelSelector) return <LabelSelector onLabelClick={handleTagClick} />;
    if (isLabelPage && !LabelSelector) return <span className='uppercase font-medium'>Owned</span>;
};

export default TagSelector;
