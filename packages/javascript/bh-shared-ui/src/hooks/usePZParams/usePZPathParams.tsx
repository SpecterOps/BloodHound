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

import { useCallback, useMemo } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import {
    certificationsPath,
    detailsPath,
    historyPath,
    labelsPath,
    objectsPath,
    privilegeZonesPath,
    rulesPath,
    savePath,
    summaryPath,
    zonesPath,
} from '../../routes';

export const usePZPathParams = () => {
    const location = useLocation();
    const { zoneId = '', labelId, ruleId, memberId } = useParams();

    const derivedValues = useMemo(() => {
        const tagId = labelId === undefined ? zoneId : labelId;

        const hasLabelId = labelId !== undefined;
        const hasZoneId = zoneId !== '';

        const isPrivilegeZonesPage = location.pathname.includes(privilegeZonesPath);
        const isCertificationsPage = location.pathname.includes(certificationsPath);
        const isSummaryPage = location.pathname.includes(summaryPath);
        const isDetailsPage = location.pathname.includes(detailsPath);
        const isHistoryPage = location.pathname.includes(historyPath);
        const isLabelPage = location.pathname.includes(`/${privilegeZonesPath}/${labelsPath}`);
        const isZonePage = location.pathname.includes(`/${privilegeZonesPath}/${zonesPath}`);

        const tagType: 'labels' | 'zones' = isLabelPage ? 'labels' : 'zones';
        const tagTypeDisplay: 'Label' | 'Zone' = isLabelPage ? 'Label' : 'Zone';
        const tagTypeDisplayPlural: 'Labels' | 'Zones' = isLabelPage ? 'Labels' : 'Zones';

        return {
            tagId,
            hasLabelId,
            hasZoneId,
            isPrivilegeZonesPage,
            isCertificationsPage,
            isSummaryPage,
            isDetailsPage,
            isHistoryPage,
            isLabelPage,
            isZonePage,
            tagType,
            tagTypeDisplay,
            tagTypeDisplayPlural,
        };
    }, [location.pathname, zoneId, labelId]);

    const { tagType } = derivedValues;

    const tagEditLink = useCallback(
        (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${savePath}`,
        [tagType]
    );
    const ruleEditLink = useCallback(
        (tagId: number | string, ruleId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${ruleId}/${savePath}`,
        [tagType]
    );
    const tagCreateLink = useCallback(
        (type?: typeof tagType) => `/${privilegeZonesPath}/${type ?? tagType}/${savePath}`,
        [tagType]
    );
    const ruleCreateLink = useCallback(
        (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${savePath}`,
        [tagType]
    );

    const tagSummaryLink = useCallback(
        (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${summaryPath}`,
        [tagType]
    );

    const tagDetailsLink = useCallback(
        (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${detailsPath}`,
        [tagType]
    );
    const ruleDetailsLink = useCallback(
        (tagId: number | string, ruleId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${ruleId}/${detailsPath}`,
        [tagType]
    );
    const objectDetailsLink = useCallback(
        (tagId: number | string, objectId: number | string, ruleId?: number | string, type?: typeof tagType) =>
            ruleId
                ? `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${ruleId}/${objectsPath}/${objectId}/${detailsPath}`
                : `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${objectsPath}/${objectId}/${detailsPath}`,
        [tagType]
    );

    return {
        ...derivedValues,
        zoneId,
        labelId,
        ruleId,
        memberId,
        tagEditLink,
        tagCreateLink,
        ruleCreateLink,
        ruleEditLink,
        tagSummaryLink,
        tagDetailsLink,
        ruleDetailsLink,
        objectDetailsLink,
    };
};
