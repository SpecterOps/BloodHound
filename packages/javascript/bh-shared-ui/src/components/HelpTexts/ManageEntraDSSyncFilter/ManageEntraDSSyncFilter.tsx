// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// SPDX-License-Identifier: Apache-2.0

import { Box, Link } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';

const General: FC = () => (
    <Typography variant='body2'>
        ManageEntraDSSyncFilter means the Domain Controller Services service principal can assign groups to the filtered
        synchronization scope of the correlated managed domain. BloodHound emits it only for application ID
        2565bd9d-da50-47d4-8b85-4c97f669dc36 when filteredSync is Enabled and syncScope is All.
    </Typography>
);

const Abuse: FC = () => (
    <Typography variant='body2' component='div'>
        <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
            <li>
                Select an attacker-controlled Entra <strong>security group</strong> containing an attacker-controlled
                user as a direct member. Do not use a direct user app-role assignment or a nested group; Entra DS does
                not honor either case for this synchronization path.
            </li>
            <li>
                In the target tenant, find the service principal with{' '}
                <code>appId=2565bd9d-da50-47d4-8b85-4c97f669dc36</code>. Read its <code>appRoles</code> and select the
                enabled role whose display name is <code>User</code>.
            </li>
            <li>
                Using the user or other directory context that controls the source service principal, create the group
                entitlement with Microsoft Graph:
                <Typography component={'pre'} variant='body2'>
                    {
                        'POST https://graph.microsoft.com/v1.0/servicePrincipals/{service-principal-object-id}/appRoleAssignedTo\nContent-Type: application/json\n\n{\n  "principalId": "{attacker-group-object-id}",\n  "resourceId": "{service-principal-object-id}",\n  "appRoleId": "{user-app-role-id}"\n}'
                    }
                </Typography>
            </li>
            <li>
                Wait for Entra DS to add the group to <code>CN=ScopedGroups,OU=AADDSSyncState,...</code> and synchronize
                the group and its direct user members. If the controlled user is cloud-only and lacks usable legacy
                password material, change its password while Entra DS is active and wait for the password to
                synchronize.
            </li>
        </ol>
    </Typography>
);

const Opsec: FC = () => (
    <Typography variant='body2'>
        App-role assignments and group membership changes generate Microsoft Entra audit activity. Subsequent use of the
        synchronized identity can generate managed-domain authentication events.
    </Typography>
);

const References: FC = () => (
    <Box className='overflow-x-auto'>
        <Link
            target='_blank'
            rel='noopener noreferrer'
            href='https://learn.microsoft.com/en-us/entra/identity/domain-services/scoped-synchronization'>
            Configure scoped synchronization
        </Link>
    </Box>
);

const ManageEntraDSSyncFilter = {
    general: General,
    windowsAbuse: Abuse,
    linuxAbuse: Abuse,
    opsec: Opsec,
    references: References,
};

export default ManageEntraDSSyncFilter;
