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

import { FC } from 'react';
import { Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body1'>Lateral movement</Typography>
            <Typography variant='body2'>
                There are several ways to pivot to a Windows system. If using Cobalt Strike's beacon, check the help
                info for the commands "psexec", "psexec_psh", "wmi", and "winrm". With Empire, consider the modules for
                Invoke-PsExec, Invoke-DCOM, and Invoke-SMBExec. With Metasploit, consider the modules
                "exploit/windows/smb/psexec", "exploit/windows/winrm/winrm_script_exec", and
                "exploit/windows/local/ps_wmi_exec". With Impacket, consider the example scripts
                psexec/wmiexec/smbexec/atexec/dcomexec. There are other alternatives like evil-winrm and crackmapexec.
                Additionally, there are several manual methods for remotely executing code on the machine, including via
                RDP, with the service control binary and interaction with the remote machine's service control manager,
                and remotely instantiating DCOM objects. For more information about these lateral movement techniques,
                see the References tab.
            </Typography>

            <Typography variant='body1'>Gathering credentials</Typography>
            <Typography variant='body2'>
                The most well-known tool for gathering credentials from a Windows system is mimikatz. mimikatz is built
                into several agents and toolsets, including Cobalt Strike's beacon, Empire, and Meterpreter. While
                running in a high integrity process with SeDebugPrivilege, execute one or more of mimikatz's credential
                gathering techniques (e.g.: sekurlsa::wdigest, sekurlsa::logonpasswords, etc.), then parse or
                investigate the output to find clear-text credentials for other users logged onto the system.
            </Typography>
            <Typography variant='body2'>
                You may also gather credentials when a user types them or copies them to their clipboard! Several
                keylogging capabilities exist, several agents and toolsets have them built-in. For instance, you may use
                meterpreter's "keyscan_start" command to start keylogging a user, then "keyscan_dump" to return the
                captured keystrokes. Or, you may use PowerSploit's Invoke-ClipboardMonitor to periodically gather the
                contents of the user's clipboard.
            </Typography>

            <Typography variant='body1'>Token Impersonation</Typography>
            <Typography variant='body2'>
                You may run into a situation where a user is logged onto the system, but you can't gather that user's
                credential. This may be caused by a host-based security product, lsass protection, etc. In those
                circumstances, you may abuse Windows' token model in several ways. First, you may inject your agent into
                that user's process, which will give you a process token as that user, which you can then use to
                authenticate to other systems on the network. Or, you may steal a process token from a remote process
                and start a thread in your agent's process with that user's token. For more information about token
                abuses, see the References tab.
            </Typography>

            <Typography variant='body1'>Disabling host-based security controls</Typography>

            <Typography variant='body2'>
                Several host-based controls may affect your ability to execute certain techniques, such as credential
                theft, process injection, command line execution, and writing files to disk. Administrators can often
                disable these host-based controls in various ways, such as stopping or otherwise disabling a service,
                unloading a driver, or making registry key changes. For more information, see the References tab.
            </Typography>
        </>
    );
};

export default Abuse;
