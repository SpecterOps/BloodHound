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

import { faArrowDown, faInbox } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, useTheme } from '@mui/material';
import { useState, DragEvent, MouseEvent, useRef, ChangeEvent } from 'react';

const FileDrop: React.FC<{
    onDrop: (files: any) => void;
}> = ({ onDrop }) => {
    const theme = useTheme();
    const inputRef = useRef<HTMLInputElement>(null);
    const [isDragActive, setDragActive] = useState(false);
    const [isHoverActive, setHoverActive] = useState(false);

    const handleClick = (_e: MouseEvent) => {
        if (inputRef.current) inputRef.current.click();
    };

    const handleChange = (_e: ChangeEvent) => onDrop(inputRef.current?.files);

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

    const handleMouseEnter = (_e: MouseEvent) => setHoverActive(true);
    const handleMouseLeave = (_e: MouseEvent) => setHoverActive(false);

    return (
        <Box
            height={300}
            borderRadius={1}
            border={2}
            px={20}
            position='relative'
            display='flex'
            flexDirection='column'
            alignItems='center'
            justifyContent='center'
            bgcolor={isHoverActive || isDragActive ? theme.palette.grey[300] : theme.palette.grey[200]}
            color={theme.palette.grey[600]}
            borderColor={theme.palette.grey[400]}
            fontWeight='bold'
            textAlign='center'>
            <input
                data-testid='ingest-file-upload'
                ref={inputRef}
                type='file'
                multiple={true}
                onChange={handleChange}
                hidden
            />
            <FontAwesomeIcon icon={isDragActive ? faArrowDown : faInbox} size='3x' />
            <p>Click here or drag and drop to upload files</p>

            <Box
                position='absolute'
                width='100%'
                height='100%'
                onClick={handleClick}
                onDragEnter={handleDragEnter}
                onDragLeave={handleDragLeave}
                onDragOver={handleDragOver}
                onMouseEnter={handleMouseEnter}
                onMouseLeave={handleMouseLeave}
                onDrop={handleDrop}
            />
        </Box>
    );
};

export default FileDrop;
