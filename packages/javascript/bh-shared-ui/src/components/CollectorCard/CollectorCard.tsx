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

import { faDownload } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, Typography } from 'doodle-ui';
import { CommunityCollectorType } from 'js-client-library';

interface CollectorCardProps {
    collectorType: CommunityCollectorType;
    version: string;
    checksum: string;
    onClickDownload?: (collectorType: CommunityCollectorType, version: string) => void;
    onClickDownloadChecksum?: (collectorType: CommunityCollectorType, version: string) => void;
    isLatest?: boolean;
    isDeprecated?: boolean;
}

const COLLECTOR_TYPE: Record<CommunityCollectorType, string> = {
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
        <div className='rounded bg-neutral-1 border border-solid border-neutral-light-4 dark:border-neutral-900'>
            <div className='p-4 flex justify-between flex-wrap gap-y-4'>
                <div className='overflow-hidden'>
                    <Typography variant='h3'>
                        {`${COLLECTOR_TYPE[collectorType]} ${version} ${getLabel(isLatest, isDeprecated)}`.trim()}
                    </Typography>
                    <Typography variant='body1'>
                        {'SHA-256: '}
                        <button
                            type='button'
                            onClick={handleOnClickDownloadChecksum}
                            title='Download Checksum'
                            className='align-baseline text-primary underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2'>
                            {checksum}
                        </button>
                    </Typography>
                </div>
                <div>
                    <Button
                        aria-label={`Download ${COLLECTOR_TYPE[collectorType]} ${version} (.zip)`}
                        variant='secondary'
                        onClick={handleOnClickDownload}>
                        <FontAwesomeIcon
                            aria-hidden='true'
                            size='lg'
                            icon={faDownload}
                            fixedWidth
                            style={{ marginRight: '8px', marginLeft: '-4px' }}
                        />{' '}
                        {`Download ${COLLECTOR_TYPE[collectorType]} ${version} (.zip)`}
                    </Button>
                </div>
            </div>
        </div>
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
