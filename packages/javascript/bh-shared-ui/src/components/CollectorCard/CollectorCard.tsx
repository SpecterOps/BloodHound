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

import { Box, Button, Link, Paper, Typography } from '@mui/material';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faDownload } from '@fortawesome/free-solid-svg-icons';

interface CollectorCardProps {
    collectorType: 'sharphound' | 'azurehound';
    version: string;
    checksum: string;
    onClickDownload?: (collectorType: 'sharphound' | 'azurehound', version: string) => void;
    onClickDownloadChecksum?: (collectorType: 'sharphound' | 'azurehound', version: string) => void;
    isLatest?: boolean;
    isDeprecated?: boolean;
}

const COLLECTOR_TYPE = {
    sharphound: 'SharpHound',
    azurehound: 'AzureHound',
};

const CollectorCard: React.FC<CollectorCardProps> = ({
    collectorType,
    version,
    checksum,
    onClickDownload = () => {},
    onClickDownloadChecksum = () => {},
    isLatest = false,
    isDeprecated = false,
}) => {
    const handleOnClickDownload = () => {
        onClickDownload(collectorType, version);
    };

    const handleOnClickDownloadChecksum = () => {
        onClickDownloadChecksum(collectorType, version);
    };

    return (
        <Paper>
            <Box p={2} display='flex' justifyContent='space-between' flexWrap='wrap' style={{ rowGap: '1rem' }}>
                <Box overflow='hidden'>
                    <Typography variant='h6'>
                        {`${COLLECTOR_TYPE[collectorType]} ${version} ${getLabel(isLatest, isDeprecated)}`.trim()}
                    </Typography>
                    <Typography variant='body1'>
                        {'SHA-256: '}
                        <Link
                            component='button'
                            variant='body1'
                            onClick={handleOnClickDownloadChecksum}
                            title='Download Checksum'
                            style={{ verticalAlign: 'baseline' }}>
                            {checksum}
                        </Link>
                    </Typography>
                </Box>
                <Box>
                    <Button
                        aria-label={`Download ${COLLECTOR_TYPE[collectorType]} ${version} (.zip)`}
                        variant='outlined'
                        color='primary'
                        onClick={handleOnClickDownload}
                        startIcon={<FontAwesomeIcon aria-hidden='true' icon={faDownload} fixedWidth />}>
                        {`Download ${COLLECTOR_TYPE[collectorType]} ${version} (.zip)`}
                    </Button>
                </Box>
            </Box>
        </Paper>
    );
};

const getLabel = (isLatest: boolean, isDeprecated: boolean): string => {
    if (isDeprecated) {
        return '(Deprecated)';
    } else if (isLatest) {
        return '(Latest)';
    } else {
        return '';
    }
};

export default CollectorCard;
