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

import { faRefresh, faTimes } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton } from '@mui/material';
import { cn } from '../../utils';
import { FileForIngest, FileStatus } from '../FileUploadDialog/types';

const FileStatusListItem: React.FC<{
    file: FileForIngest;
    onRemove: () => void;
    onRefresh: (file: FileForIngest) => void;
    percentCompleted: number;
}> = ({ file, onRemove, onRefresh, percentCompleted = 0 }) => {
    const percentWithFallback = file.status === FileStatus.DONE ? 100 : percentCompleted;
    const hasErrors = !!file?.errors?.length || file.status === FileStatus.FAILURE;
    const clampedPercent = Math.max(0, Math.min(100, Math.round(percentWithFallback ?? 0)));
    const shouldBeFullWidth = hasErrors || [FileStatus.DONE, FileStatus.FAILURE].includes(file.status);
    const progressBarWidth = shouldBeFullWidth ? '100%' : `${percentCompleted}%`;

    return (
        <div className='mb-2 relative flex flex-row h-8 justify-between text-sm'>
            <div className='pl-3 flex items-center z-10'>
                <span className='pr-2'>{file.file.name}</span>{' '}
                {!!percentWithFallback && !hasErrors && <span>{clampedPercent}%</span>}
                {hasErrors && <span className='text-error'>Failed to Upload</span>}
            </div>
            <div
                className={cn('absolute h-8 rounded-lg transition-all', {
                    'bg-purple-300 opacity-20': !hasErrors,
                    'bg-red-500 opacity-10': hasErrors,
                })}
                style={{ maxWidth: '600px', width: progressBarWidth }}
            />

            <div>
                {file.status === FileStatus.READY && (
                    <IconButton
                        onClick={onRemove}
                        aria-label='Remove item'
                        className='hover:bg-slate-400 rounded-sm w-4 h-3 m-2 justify-self-end'>
                        <FontAwesomeIcon size='xs' icon={faTimes} />
                    </IconButton>
                )}
                {file.status === FileStatus.FAILURE && (
                    <IconButton
                        onClick={() => onRefresh(file)}
                        aria-label='Retry upload'
                        className='hover:bg-slate-400 rounded-sm w-4 h-3 m-2 justify-self-end'>
                        <FontAwesomeIcon size='xs' icon={faRefresh} />
                    </IconButton>
                )}
            </div>
        </div>
    );
};

export default FileStatusListItem;
