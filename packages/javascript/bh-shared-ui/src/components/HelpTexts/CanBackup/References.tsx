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

import { Box, Link } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-groups#backup-operators'>
                Active Directory security groups (Microsoft)
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://www.thehacker.recipes/ad/movement/builtins/security-groups'>
                Security groups (The Hacker Recipes)
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://raw.githubusercontent.com/Wh04m1001/Random/main/BackupOperators.cpp'>
                BackupOperators.cpp (Filip Dragovic)
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://github.com/improsec/BackupOperatorToolkit'>
                BackupOperatorToolkit (Improsec)
            </Link>            
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/windows/win32/api/winreg/nf-winreg-regcreatekeyexa'>
                REG_OPTION_BACKUP_RESTORE - RegCreateKeyExA (Microsoft)
            </Link>             
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e8fb45c1-a03d-44ca-b7ae-47385cfd7997'>
                FILE_OPEN_FOR_BACKUP_INTENT - SMB2 Create (Microsoft)
            </Link>            
        </Box>
    );
};

export default References;
