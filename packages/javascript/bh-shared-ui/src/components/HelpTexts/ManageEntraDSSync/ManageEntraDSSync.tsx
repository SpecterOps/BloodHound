// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// SPDX-License-Identifier: Apache-2.0

import { Box, Link } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';
import Composition from './Composition';

const General: FC = () => (
    <Typography variant='body2'>
        ManageEntraDSSync means the source can broaden the synchronization boundary of the correlated Microsoft Entra
        Domain Services (Entra DS) domain and cause eligible Entra users to materialize with Domain Users access. It is
        emitted from an AZManageEntraDS principal regardless of the current filteredSync and syncScope settings.
    </Typography>
);

const Abuse: FC = () => (
    <Typography variant='body2' component='div'>
        <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
            <li>
                Read the managed domain's current <code>filteredSync</code> and <code>syncScope</code> values. If{' '}
                <code>filteredSync=Enabled</code>, add an Entra security group of which the attacker-controlled user is
                a direct member to the filter, as described for ManageEntraDSSyncFilter. If{' '}
                <code>filteredSync=Disabled</code> but <code>syncScope=CloudOnly</code>, change <code>syncScope</code>{' '}
                to <code>All</code> with the ARM PUT workflow described for AZManageEntraDS.
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
