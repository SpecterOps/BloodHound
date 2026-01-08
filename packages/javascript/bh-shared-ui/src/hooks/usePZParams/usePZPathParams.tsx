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

    const tagEditLink = (tagId: number | string, type?: typeof tagType) =>
        `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${savePath}`;
    const ruleEditLink = (tagId: number | string, ruleId: number | string, type?: typeof tagType) =>
        `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${ruleId}/${savePath}`;
    const tagCreateLink = (type?: typeof tagType) => `/${privilegeZonesPath}/${type ?? tagType}/${savePath}`;
    const ruleCreateLink = (tagId: number | string, type?: typeof tagType) =>
        `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${rulesPath}/${savePath}`;

    const tagSummaryLink = (tagId: number | string, type?: typeof tagType) =>
        `/${privilegeZonesPath}/${type ?? tagType}/${tagId}/${summaryPath}?environmentAggregation=all`;

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

    return {
        tagId,
        zoneId,
        labelId,
        ruleId,
        memberId,
        hasLabelId,
        hasZoneId,
        isLabelPage,
        isZonePage,
        isPrivilegeZonesPage,
        isCertificationsPage,
        isHistoryPage,
        isSummaryPage,
        isDetailsPage,
        tagType,
        tagTypeDisplay,
        tagTypeDisplayPlural,
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
