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

import capitalize from 'lodash/capitalize';
import { useLocation } from 'react-router-dom';
import {
    useHighestPrivilegeTagId,
    useOwnedTagId,
    usePrivilegeZoneAnalysis,
    useZonePathParams,
} from '../../../../hooks';
import { detailsPath, labelPath, privilegeZonesPath, savePath, selectorPath, zonePath } from '../../../../routes';
import { useAppNavigate } from '../../../../utils';

export const useTagFormUtils = () => {
    const navigate = useAppNavigate();
    const location = useLocation();

    const { tagId, zoneId, labelId } = useZonePathParams();

    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const ownedId = useOwnedTagId();

    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();

    const isLabelLocation = location.pathname.includes(`/${privilegeZonesPath}/${labelPath}`);
    const isZoneLocation = location.pathname.includes(`/${privilegeZonesPath}/${zonePath}`);

    const tagType: 'label' | 'zone' = isLabelLocation ? 'label' : 'zone';
    const tagTypeDisplay: 'Label' | 'Zone' = capitalize(tagType) as 'Label' | 'Zone';

    const isUpdateZoneLocation = isZoneLocation && zoneId !== '';
    const isUpdateLabelLocation = isLabelLocation && labelId;

    const handleCreateNavigate = (tagId: number) => {
        navigate(`${location.pathname}/${tagId}`, { replace: true });
        navigate(`/${privilegeZonesPath}/${tagType}/${tagId}/${selectorPath}/${savePath}`);
    };

    const handleUpdateNavigate = () => navigate(`/${privilegeZonesPath}/${tagType}/${tagId}/${detailsPath}`);

    const handleDeleteNavigate = () =>
        navigate(`/${privilegeZonesPath}/${tagType}/${tagType === 'zone' ? topTagId : ownedId}/${detailsPath}`);

    const showAnalysisToggle = privilegeZoneAnalysisEnabled && isUpdateZoneLocation && zoneId !== topTagId?.toString();

    const showDeleteButton = (): boolean => {
        if (zoneId === '' && !labelId) return false;
        if (labelId === ownedId?.toString()) return false;
        if (zoneId === topTagId?.toString()) return false;
        return true;
    };

    const formTitleFromPath = (): string => {
        if (isLabelLocation && !labelId) return 'Create new Label';
        if (isZoneLocation && zoneId === '') return 'Create new Zone';
        if (isUpdateLabelLocation) return 'Edit Label Details';
        if (isUpdateZoneLocation) return 'Edit Zone Details';
        return 'Tag Details';
    };

    const formTitle = formTitleFromPath();

    const disableNameInput = tagId === topTagId?.toString() || tagId === ownedId?.toString();

    return {
        privilegeZoneAnalysisEnabled,
        disableNameInput,
        formTitle,
        tagType,
        tagTypeDisplay,
        isLabelLocation,
        isZoneLocation,
        isUpdateZoneLocation,
        showAnalysisToggle,
        showDeleteButton,
        handleCreateNavigate,
        handleUpdateNavigate,
        handleDeleteNavigate,
    };
};
