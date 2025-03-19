// Copyright 2023 Specter Ops, Inc.
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

import { Box, useTheme } from '@mui/material';
import CollectorCard, { LabelType } from '../CollectorCard';

interface CollectorDownloadFile {
    fileName: string;
    os: string;
    arch: string;
    onClickDownload: () => void;
    onClickDownloadChecksum: () => void;
}

interface CollectorCardProps {
    collectorType: 'sharphound' | 'azurehound';
    version: string;
    timestamp: number;
    downloadArtifacts: CollectorDownloadFile[];
    label?: LabelType;
    isPrerelease?: boolean;
}

interface CollectorCardListProps {
    collectors: CollectorCardProps[];
}

const CollectorCardList: React.FC<CollectorCardListProps> = ({ collectors }) => {
    const theme = useTheme();

    return (
        <Box display='grid' rowGap={theme.spacing(2)}>
            {collectors.map((collector, index) => (
                <Box key={index}>
                    <CollectorCard
                        collectorType={collector.collectorType}
                        version={collector.version}
                        timestamp={collector.timestamp}
                        label={collector.label}
                        isPrerelease={collector.isPrerelease}
                        downloadArtifacts={collector.downloadArtifacts}
                    />
                </Box>
            ))}
        </Box>
    );
};

export default CollectorCardList;
