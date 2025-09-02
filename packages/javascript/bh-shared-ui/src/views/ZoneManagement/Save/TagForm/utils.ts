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
import { ROUTE_ZONE_MANAGEMENT_ROOT } from '../../../../routes';
import { useAppNavigate } from '../../../../utils';

export const useTagFormUtils = () => {
    const navigate = useAppNavigate();
    const location = useLocation();

    const { tagId, tierId, labelId } = useZonePathParams();

    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const ownedId = useOwnedTagId();

    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();

    const isLabelLocation = location.pathname.includes(`${ROUTE_ZONE_MANAGEMENT_ROOT}/save/label`);
    const isTierLocation = location.pathname.includes(`${ROUTE_ZONE_MANAGEMENT_ROOT}/save/tier`);

    const tagKind: 'label' | 'tier' = isLabelLocation ? 'label' : 'tier';
    const tagKindDisplay: 'Label' | 'Tier' = capitalize(tagKind) as 'Label' | 'Tier';

    const isUpdateTierLocation = isTierLocation && tierId !== '';
    const isUpdateLabelLocation = isLabelLocation && labelId;

    const handleCreateNavigate = (tagId: number) => {
        navigate(`${location.pathname}/${tagId}`, { replace: true });
        navigate(`${location.pathname}/${tagId}/selector`);
    };

    const handleUpdateNavigate = () => navigate(`${ROUTE_ZONE_MANAGEMENT_ROOT}/${tagKind}/${tagId}/details`);

    const handleDeleteNavigate = () =>
        navigate(`${ROUTE_ZONE_MANAGEMENT_ROOT}/${tagKind}/${tagKind === 'tier' ? topTagId : ownedId}/details`);

    const showAnalysisToggle = privilegeZoneAnalysisEnabled && isUpdateTierLocation && tierId !== topTagId?.toString();

    const showDeleteButton = (): boolean => {
        if (tierId === '' && !labelId) return false;
        if (labelId === ownedId?.toString()) return false;
        if (tierId === topTagId?.toString()) return false;
        return true;
    };

    const formTitleFromPath = (): string => {
        if (isLabelLocation && !labelId) return 'Create new Label';
        if (isTierLocation && tierId === '') return 'Create new Tier';
        if (isUpdateLabelLocation) return 'Edit Label Details';
        if (isUpdateTierLocation) return 'Edit Tier Details';
        return 'Tag Details';
    };

    const formTitle = formTitleFromPath();

    const disableNameInput = tagId === topTagId?.toString() || tagId === ownedId?.toString();

    return {
        privilegeZoneAnalysisEnabled,
        disableNameInput,
        formTitle,
        tagKind,
        tagKindDisplay,
        isLabelLocation,
        isTierLocation,
        isUpdateTierLocation,
        showAnalysisToggle,
        showDeleteButton,
        handleCreateNavigate,
        handleUpdateNavigate,
        handleDeleteNavigate,
    };
};
