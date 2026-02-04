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

import { faArrowDown, faInbox, IconDefinition } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { DragEvent, useRef, useState } from 'react';
import { adaptClickHandlerToKeyDown, cn } from '../../utils';

const FileDrop: React.FC<{
    onDrop: (files: any) => void;
    disabled: boolean;
    accept?: string[];
    multiple?: boolean;
    icon?: IconDefinition;
}> = ({ onDrop, disabled, accept, multiple = true, icon = faInbox }) => {
    const inputRef = useRef<HTMLInputElement>(null);
    const [isDragActive, setDragActive] = useState(false);
    const [isHoverActive, setHoverActive] = useState(false);

    const handleClick = () => {
        if (inputRef.current) inputRef.current.click();
    };

    const handleChange = () => onDrop(inputRef.current?.files);

    const handleDrop = (e: DragEvent) => {
        e.preventDefault();
        onDrop(e.dataTransfer.files);
        setDragActive(false);
    };

    const handleDragEnter = (e: DragEvent) => {
        e.preventDefault();
        setDragActive(true);
    };

    const handleDragLeave = (e: DragEvent) => {
        e.preventDefault();
        setDragActive(false);
    };

    const handleDragOver = (e: DragEvent) => e.preventDefault();

    const handleMouseEnter = () => setHoverActive(true);
    const handleMouseLeave = () => setHoverActive(false);

    const formatAcceptList = () => (accept && accept.length ? accept.join(',') : undefined);

    return (
        <div
            className={cn(
                'cursor-pointer h-80 rounded font-bold text-center border-2 border-contrast px-32 relative flex flex-col items-center justify-center bg-neutral-2',
                {
                    'cursor-default opacity-50': disabled,
                    'bg-neutral-3': isHoverActive || isDragActive || disabled,
                }
            )}>
            <input
                data-testid='ingest-file-upload'
                disabled={disabled}
                ref={inputRef}
                type='file'
                multiple={multiple}
                onChange={handleChange}
                hidden
                accept={formatAcceptList()}
            />
            <FontAwesomeIcon icon={isDragActive ? faArrowDown : icon} size='3x' />
            <p className='pt-2'>
                {multiple
                    ? 'Click here or drag and drop to upload JSON or zip/compressed JSON files'
                    : 'Click here or drag and drop to upload a JSON file'}
            </p>
            <div
                role='button'
                tabIndex={0}
                className='absolute size-full'
                onClick={handleClick}
                onDragEnter={handleDragEnter}
                onDragLeave={handleDragLeave}
                onDragOver={handleDragOver}
                onMouseEnter={handleMouseEnter}
                onMouseLeave={handleMouseLeave}
                onKeyDown={adaptClickHandlerToKeyDown(handleClick)}
                onDrop={handleDrop}></div>
        </div>
    );
};

export default FileDrop;
