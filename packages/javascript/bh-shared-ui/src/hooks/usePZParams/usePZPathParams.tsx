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
    privilegeZonesPath,
    summaryPath,
    zonesPath,
} from '../../routes';

export const usePZPathParams = () => {
    const location = useLocation();
    const { zoneId = '', labelId, selectorId, memberId } = useParams();
    const tagId = labelId === undefined ? zoneId : labelId;

    const hasLabelId = labelId !== undefined;
    const hasZoneId = zoneId !== '';

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
        zoneId,
        labelId,
        selectorId,
        memberId,
        hasLabelId,
        hasZoneId,
        isLabelPage,
        isCertificationsPage,
        isHistoryPage,
        isSummaryPage,
        isDetailsPage,
        isZonePage,
        tagType,
        tagTypeDisplay,
        tagTypeDisplayPlural,
    };
};