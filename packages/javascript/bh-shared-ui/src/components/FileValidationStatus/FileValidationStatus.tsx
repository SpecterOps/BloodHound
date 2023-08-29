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

import { faCheck, faSync, faTimes } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Avatar, Box } from '@mui/material';
import { FileForIngest, FileStatus } from '../FileUploadDialog/types';

const FileValidationStatus: React.FC<{ file: FileForIngest }> = ({ file }) => {
    if (file.errors?.length) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Avatar sx={{ bgcolor: 'red', width: 24, height: 24 }}>
                    <FontAwesomeIcon icon={faTimes} size='xs' color='white' />
                </Avatar>
                <div>
                    {file.errors.map((error, i) => (
                        <div key={i}>{error}</div>
                    ))}
                </div>
            </Box>
        );
    }
    if (file.status === FileStatus.UPLOADING) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Box
                    sx={{
                        animation: 'spin 2s linear infinite',
                        '@keyframes spin': {
                            '0%': {
                                transform: 'rotate(0deg)',
                            },
                            '100%': {
                                transform: 'rotate(360deg)',
                            },
                        },
                    }}>
                    <FontAwesomeIcon icon={faSync} size='sm' color='grey' />
                </Box>
                <div>Uploading...</div>
            </Box>
        );
    }
    return (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Avatar sx={{ bgcolor: 'green', width: 24, height: 24 }}>
                <FontAwesomeIcon icon={faCheck} transform='shrink-1' size='xs' color='white' />
            </Avatar>
            <div>{file.status === FileStatus.READY ? 'Ready' : 'Done'}</div>
        </Box>
    );
};

export default FileValidationStatus;
