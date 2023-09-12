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
import { Grid, IconButton } from '@mui/material';
import { FileForIngest, FileStatus } from '../FileUploadDialog/types';
import FileValidationStatus from '../FileValidationStatus';

const FileStatusListItem: React.FC<{
    file: FileForIngest;
    onRemove: () => void;
}> = ({ file, onRemove }) => {
    return (
        <Grid
            container
            width='100%'
            minHeight={32}
            fontSize={12}
            borderLeft={1}
            borderRight={1}
            borderBottom={1}
            borderColor='lightgray'>
            <Grid item xs={6} sx={{ display: 'flex', alignItems: 'center', paddingLeft: '4px' }}>
                {file.file.name}
            </Grid>
            <Grid item xs={5} sx={{ display: 'flex', alignItems: 'center' }}>
                <FileValidationStatus file={file} />
            </Grid>
            {file.status === FileStatus.READY && (
                <Grid item xs={1} sx={{ display: 'flex', alignItems: 'center', justifyContent: 'end' }}>
                    <IconButton
                        onClick={onRemove}
                        sx={{
                            '&:hover': {
                                backgroundColor: 'lightgray',
                            },
                            borderRadius: '2px',
                            width: 28,
                            height: 28,
                            margin: '2px',
                        }}>
                        <FontAwesomeIcon size='xs' icon={faTimes} />
                    </IconButton>
                </Grid>
            )}
        </Grid>
    );
};

export default FileStatusListItem;
