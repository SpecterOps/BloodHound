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
            EntraDSFor is a non-traversable, post-processed correlation from a Microsoft Entra Domain Services (Entra
            DS) resource to its managed AD Domain. BloodHound requires a unique normalized match between{' '}
            <code>AZEntraDS.domainname</code> and <code>Domain.name</code>, then corroborates the domain SID through the
            tenant's synchronized AAD DC Administrators group.
        </Typography>
        <Typography variant='body2'>
            The group must be represented by an AZGroup named AAD DC Administrators, a SyncedToEntraDSGroup relationship
            to its Entra DS Group, and a matching <code>domainsid</code> on the candidate Domain. BloodHound fails
            closed when the name match is missing or ambiguous or the group correlation does not corroborate the SID.
        </Typography>
    </>
);

const References: FC = () => (
    <Box className='overflow-x-auto'>
        <Link
            target='_blank'
            rel='noopener noreferrer'
            href='https://learn.microsoft.com/en-us/entra/identity/domain-services/synchronization'>
            How objects and credentials are synchronized in a Microsoft Entra Domain Services managed domain
        </Link>
        <br />
        <Link
            target='_blank'
            rel='noopener noreferrer'
            href='https://learn.microsoft.com/en-us/entra/identity/domain-services/tutorial-create-instance-advanced#configure-an-administrative-group'>
            Configure the AAD DC Administrators group
        </Link>
    </Box>
);

const EntraDSFor = { general: General, references: References };

export default EntraDSFor;
