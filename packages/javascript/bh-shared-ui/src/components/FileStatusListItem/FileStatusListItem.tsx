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

import { faTimes } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton } from '@mui/material';
import { cn } from '../../utils';
import { FileForIngest, FileStatus } from '../FileUploadDialog/types';

const FileStatusListItem: React.FC<{
    file: FileForIngest;
    onRemove: () => void;
    percentCompleted: number;
}> = ({ file, onRemove, percentCompleted }) => {
    const hasErrors = !!file?.errors?.length;
    const clampedPercent = Math.max(0, Math.min(100, Math.round(percentCompleted ?? 0)));
    const progressBarWidth = hasErrors ? '100%' : `${percentCompleted}%`;

    return (
        <div className='mb-2 relative flex flex-row h-8 justify-between'>
            <div
                className={cn('absolute h-8 opacity-40 rounded-lg transition-all', {
                    'bg-purple-300': !hasErrors,
                    'bg-red-300': hasErrors,
                })}
                style={{ maxWidth: '600px', width: progressBarWidth }}
            />

            <div className='pl-3 flex items-center'>
                <span className='pr-2'>{file.file.name}</span>+{' '}
                {percentCompleted && !hasErrors && <span>{clampedPercent}%</span>}
                {hasErrors && <span className='text-error'>Failed to Upload</span>}
            </div>
            <div>
                {file.status === FileStatus.READY && (
                    <IconButton
                        onClick={onRemove}
                        className='hover:bg-slate-400 rounded-sm w-4 h-3 m-2 justify-self-end'>
                        <FontAwesomeIcon size='xs' icon={faTimes} />
                    </IconButton>
                )}
            </div>
        </div>
    );
};

export default FileStatusListItem;
