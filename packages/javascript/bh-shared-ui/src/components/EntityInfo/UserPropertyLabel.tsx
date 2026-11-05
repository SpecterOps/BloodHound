// Copyright 2026 Specter Ops, Inc.
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

import { Tooltip } from 'doodle-ui';
import { ActiveDirectoryKindProperties } from '../../graphSchema';
import { AppIcon } from '../AppIcon';

const propertyHelp: Partial<Record<ActiveDirectoryKindProperties, { description: string; location: string }>> = {
    [ActiveDirectoryKindProperties.AdminCount]: {
        description:
            'Indicates whether Active Directory has marked this account for protected-account handling. The flag can remain after privileged group membership is removed.',
        location: 'Enable View → Advanced Features, then open the user’s Properties → Attribute Editor → adminCount.',
    },
    [ActiveDirectoryKindProperties.Sensitive]: {
        description: 'Marks the account as sensitive and prevents delegation of its security context.',
        location:
            'Open the user’s Properties → Account → Account options → “Account is sensitive and cannot be delegated.”',
    },
    [ActiveDirectoryKindProperties.PasswordNotRequired]: {
        description: 'Indicates a password-length policy exemption; it does not establish that the password is blank.',
        location:
            'Enable View → Advanced Features, then open the user’s Properties → Attribute Editor → userAccountControl. Inspect the PASSWD_NOTREQD flag (0x20 / 32).',
    },
};

export const UserPropertyLabel = ({ property, label }: { property?: string; label: string }) => {
    const help =
        property && Object.hasOwn(propertyHelp, property)
            ? propertyHelp[property as ActiveDirectoryKindProperties]
            : undefined;

    if (!help) return <>{label}</>;

    return (
        <Tooltip
            contentWidth='lg'
            tooltipPosition='left'
            triggerProps={{ asChild: false, 'aria-label': `About ${label}`, className: 'text-left' }}
            tooltip={
                <>
                    <span className='block'>{help.description}</span>
                    <span className='mt-2 block'>
                        Find this in Active Directory Users and Computers: {help.location}
                    </span>
                </>
            }>
            <span className='inline-flex items-center gap-1'>
                {label}
                <AppIcon.Info size={14} className='shrink-0' aria-hidden='true' />
            </span>
        </Tooltip>
    );
};
