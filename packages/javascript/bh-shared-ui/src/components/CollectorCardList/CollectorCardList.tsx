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
import CollectorCard from '../CollectorCard';

interface CollectorCardListProps {
    collectors: {
        collectorType: 'sharphound' | 'azurehound';
        version: string;
        checksum: string;
        isLatest: boolean;
        isDeprecated: boolean;
        onClickDownload: (collectorType: 'sharphound' | 'azurehound', version: string) => void;
        onClickDownloadChecksum: (collectorType: 'sharphound' | 'azurehound', version: string) => void;
    }[];
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
                        checksum={collector.checksum}
                        isLatest={collector.isLatest}
                        isDeprecated={collector.isDeprecated}
                        onClickDownload={collector.onClickDownload}
                        onClickDownloadChecksum={collector.onClickDownloadChecksum}
                    />
                </Box>
            ))}
        </Box>
    );
};

export default CollectorCardList;
