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

import { Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const Abuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                There are several ways to abuse Backup Operators to compromise a remote host. The easiest is dumping the 
                SYSTEM, SAM, and SECURITY hives from the target host, extracting the computer account credentials, and 
                then performing a Silver Ticket or Pass The Hash attack.
            </Typography>
            <Typography variant='body1'>Remotely Dump SYSTEM, SAM, and SECURITY Registry Hives</Typography>
            <Typography variant='body2'>
                This is a fairly well-known tradecraft technique. Filip Dragovich released some of the first public tooling
                for it as BackupOperators.cpp.

                Other known public tooling:
                  -  https://github.com/improsec/BackupOperatorToolkit
                  -  https://github.com/mpgn/BackupOperatorToDA
                  -  https://github.com/snovvcrash/RemoteRegSave
                  -  https://github.com/horizon3ai/backup_dc_registry
                  -  https://www.thehacker.recipes/ad/movement/credentials/dumping/sam-and-lsa-secrets
            </Typography>
            <Typography variant='body1'>Remote Registry Manipulation</Typography>
            <Typography variant='body2'>
                Backup Operators are granted a default ACE on the winreg named pipe which allows a connection to the Remote 
                Registry.  This is similar opening a connection to an SMB share. From there, a key can be opened with the 
                Reg_Option_Backup_Restore flag and enabled SeBackupPrivilege and SeRestorePrivilege to circumvent any
                security descriptor on individual registry keys or subkeys. Registry keys, subkeys, and values can be 
                read, written, or deleted. It is possible to modify the security descriptor of an existing registry key.
            </Typography>
            <Typography variant='body1'>SMB Share Access</Typography>
            <Typography variant='body2'>
                Backup Operators are granted a default ACE on the default Admin shares in Windows which allows them to 
                open a connection to the C$, Admin$, and IPC$ administrative shares.  The SMB Tree_Connect operation is 
                used for this part of the operation.
            </Typography>                
            <Typography variant='body2'>
                From there, any file or directory can be opened with the FILE_OPEN_FOR_BACKUP_INTENT flag along with 
                enabled SeBackupPrivilege and SeRestorePrivilege to bypass any security descriptor on any file that 
                is not locked for use. The SMB Create operation is used for this part of the operation.  Files on the 
                remote host can be created, modified, and deleted.  This includes those owned by TrustedInstaller or 
                SYSTEM.
            </Typography>
            <Typography variant='body2'>
                If a volume shadow copy already exists, files on it can be copied regardless of DACL. Shadow copies on a 
                remote host can be enumerated: https://github.com/HiraokaHyperTools/LibEnumRemotePreviousVersion or via
                the Titanis smb2client.  Without an existing shadow copy, it is not possible to open a handle to a 
                file that is locked open, such as NTDS.dit.
            </Typography>
            <Typography variant='body1'>AutoStart Persistence</Typography>
            <Typography variant='body2'>
                With the capability to remotely modify the registry and file system, just about any of the 150+ autostart 
                persistence methods should be possible.

                https://www.hexacorn.com/blog/category/autostart-persistence/
            </Typography>                        
        </>
    );
};

export default Abuse;
