// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// SPDX-License-Identifier: Apache-2.0

import { Box, Link } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';
import Composition from './Composition';

const General: FC = () => (
    <>
        <Typography variant='body2'>
            BloodHound emits ManageEntraDSSync from each principal with AZManageEntraDS to a correlated Microsoft Entra
            Domain Services (Entra DS) resource. The destination is the RID 513 Domain Users group in the Domain
            identified by EntraDSFor.
        </Typography>
        <Typography variant='body2'>
            The relationship is independent of current <code>filteredSync</code> and <code>syncScope</code> values
            because the source can change both settings. BloodHound does not emit an unconditional attack edge from the
            AZEntraDS resource.
        </Typography>
    </>
);

const Abuse: FC = () => (
    <Typography variant='body2' component='div'>
        <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
            <li>
                Read the managed domain's current synchronization settings. If the controlled user originated
                on-premises and <code>syncScope=CloudOnly</code>, first change <code>syncScope</code> to{' '}
                <code>All</code> using the ARM PUT workflow described by the AZManageEntraDS edge. If{' '}
                <code>filteredSync=Enabled</code>, then add an Entra security group containing the controlled user as a
                direct member to the filter, as described by the ManageEntraDSSyncFilter edge.
            </li>
            <li>
                Wait for the controlled user to synchronize to Entra DS. Poll for its Entra object ID in{' '}
                <code>msDS-aadObjectId</code>, or use repeated non-destructive authentication attempts when LDAP
                inspection is unavailable.
            </li>
        </ol>
    </Typography>
);

const Opsec: FC = () => (
    <Typography variant='body2'>
        Managed-domain updates are recorded in the Azure Activity Log and can trigger a full resynchronization,
        including deletion of objects that fall outside the new boundary.
    </Typography>
);

const References: FC = () => (
    <Box className='overflow-x-auto'>
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
            href='https://learn.microsoft.com/en-us/entra/identity/domain-services/scoped-synchronization'>
            Configure scoped synchronization
        </Link>
        <br />
        <Link target='_blank' rel='noopener noreferrer' href='https://attack.mitre.org/techniques/T1136/002/'>
            MITRE ATT&amp;CK T1136.002: Create Account - Domain Account
        </Link>
    </Box>
);

const ManageEntraDSSync = {
    general: General,
    windowsAbuse: Abuse,
    linuxAbuse: Abuse,
    opsec: Opsec,
    references: References,
    composition: Composition,
};

export default ManageEntraDSSync;
