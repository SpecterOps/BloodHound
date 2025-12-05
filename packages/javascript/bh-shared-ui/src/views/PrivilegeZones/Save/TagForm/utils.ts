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

import { useHighestPrivilegeTagId, useOwnedTagId, usePrivilegeZoneAnalysis, usePZPathParams } from '../../../../hooks';
import { detailsPath, privilegeZonesPath } from '../../../../routes';
import { useAppNavigate } from '../../../../utils';

export const useTagFormUtils = () => {
    const navigate = useAppNavigate();
    const {
        tagId,
        zoneId,
        labelId,
        tagType,
        tagTypeDisplay,
        isZonePage,
        isLabelPage,
        tagEditLink,
        ruleCreateLink,
        tagDetailsLink,
    } = usePZPathParams();

    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const ownedId = useOwnedTagId();

    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();

    const isUpdateZoneLocation = isZonePage && zoneId !== '';
    const isUpdateLabelLocation = isLabelPage && !!labelId;

    const handleCreateNavigate = (tagId: number) => {
        navigate(tagEditLink(tagId), { replace: true });
        navigate(ruleCreateLink(tagId));
    };

    const handleUpdateNavigate = (tagId: number | string) => navigate(tagDetailsLink(tagId));

    const handleDeleteNavigate = () =>
        navigate(`/${privilegeZonesPath}/${tagType}/${tagType === 'zones' ? topTagId : ownedId}/${detailsPath}`);

    const showAnalysisToggle = privilegeZoneAnalysisEnabled && isUpdateZoneLocation && zoneId !== topTagId?.toString();

    const showDeleteButton = (): boolean => {
        if (zoneId === '' && !labelId) return false;
        if (labelId === ownedId?.toString()) return false;
        if (zoneId === topTagId?.toString()) return false;
        return true;
    };

    const formTitleFromPath = (): string => {
        if (isLabelPage && !labelId) return 'Create new Label';
        if (isZonePage && zoneId === '') return 'Create new Zone';
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
        tagId,
        tagType,
        tagTypeDisplay,
        isLabelPage,
        isZonePage,
        isUpdateZoneLocation,
        showAnalysisToggle,
        showDeleteButton,
        handleCreateNavigate,
        handleUpdateNavigate,
        handleDeleteNavigate,
    };
};
