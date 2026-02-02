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

import { useLocation, useParams } from '@tanstack/react-router';
import { useMemo } from 'react';
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

type TagType = 'labels' | 'zones';

export const usePZPathParams = () => {
    const pathParams = useParams({ strict: false, shouldThrow: false });
    const { pathname } = useLocation();

    console.log('superfluous!');

    return useMemo(() => {
        const { zoneId = '', labelId, ruleId, objectId: memberId } = pathParams;

        const hasLabelId = labelId !== undefined;
        const hasZoneId = zoneId !== '';

        const tagId = labelId === undefined ? zoneId : labelId;

        const tagType: TagType = hasLabelId ? 'labels' : 'zones';
        const tagTypeDisplay: 'Label' | 'Zone' = hasLabelId ? 'Label' : 'Zone';
        const tagTypeDisplayPlural: Capitalize<TagType> = hasLabelId ? 'Labels' : 'Zones';

        const tagEditLink = (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${savePath}`;
        const ruleEditLink = (tagId: number | string, ruleId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${ruleId}/${savePath}`;
        const tagCreateLink = (type?: typeof tagType) => `/${privilegeZonesPath}/${type ?? tagType}/${savePath}`;
        const ruleCreateLink = (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${savePath}`;

        const tagSummaryLink = (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${summaryPath}`;

        const tagDetailsLink = (tagId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${detailsPath}`;
        const ruleDetailsLink = (tagId: number | string, ruleId: number | string, type?: typeof tagType) =>
            `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${ruleId}/${detailsPath}`;
        const objectDetailsLink = (
            tagId: number | string,
            objectId: number | string,
            ruleId?: number | string,
            type?: typeof tagType
        ) =>
            ruleId
                ? `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${ruleId}/${objectsPath}/${objectId}/${detailsPath}`
                : `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${objectsPath}/${objectId}/${detailsPath}`;

        const isPrivilegeZonesPage = pathname.includes(privilegeZonesPath);
        const isCertificationsPage = pathname.includes(certificationsPath);
        const isSummaryPage = pathname.includes(summaryPath);
        const isDetailsPage = pathname.includes(detailsPath);
        const isHistoryPage = pathname.includes(historyPath);
        const isLabelPage = pathname.includes(`/${privilegeZonesPath}/${labelsPath}`);
        const isZonePage = pathname.includes(`/${privilegeZonesPath}/${zonesPath}`);

        return {
            tagId,
            zoneId,
            labelId,
            ruleId,
            memberId,
            tagType,
            tagTypeDisplay,
            tagTypeDisplayPlural,
            hasLabelId,
            hasZoneId,
            tagEditLink,
            ruleEditLink,
            tagCreateLink,
            ruleCreateLink,
            tagSummaryLink,
            tagDetailsLink,
            ruleDetailsLink,
            objectDetailsLink,
            isPrivilegeZonesPage,
            isCertificationsPage,
            isSummaryPage,
            isDetailsPage,
            isHistoryPage,
            isLabelPage,
            isZonePage,
        };
    }, [pathParams, pathname]);
};
