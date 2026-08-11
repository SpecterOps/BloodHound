// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// SPDX-License-Identifier: Apache-2.0

import { Box, Link } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';

const General: FC = () => (
    <>
        <Typography variant='body2'>
            BloodHound emits ManageEntraDSSyncFilter from the service principal associated through AZRunsAs with Domain
            Controller Services application ID <code>2565bd9d-da50-47d4-8b85-4c97f669dc36</code>. The service principal
            and Microsoft Entra Domain Services (Entra DS) resource must belong to the same tenant, and the resource
            must have <code>filteredSync=Enabled</code> and <code>syncScope=All</code>.
        </Typography>
        <Typography variant='body2'>
            The destination is the RID 513 Domain Users group in the Domain identified by EntraDSFor.
        </Typography>
    </>
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
        <br />
        <Link
            target='_blank'
            rel='noopener noreferrer'
            href='https://learn.microsoft.com/en-us/entra/identity/domain-services/synchronization'>
            Microsoft Entra Domain Services synchronization
        </Link>
        <br />
        <Link
            target='_blank'
            rel='noopener noreferrer'
            href='https://learn.microsoft.com/en-us/graph/api/resources/approleassignment'>
            Microsoft Graph appRoleAssignedTo resource
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
