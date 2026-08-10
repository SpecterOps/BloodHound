// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// SPDX-License-Identifier: Apache-2.0

import { Typography } from 'doodle-ui';
import { FC } from 'react';

const General: FC = () => (
    <Typography variant='body2'>
        EntraDSFor is a non-traversable, post-processed correlation from an AZEntraDS resource to its managed AD Domain.
        BloodHound requires matching normalized domain names and corroborates the domain SID through the tenant's
        synchronized AAD DC Administrators group.
    </Typography>
);

const EntraDSFor = { general: General };

export default EntraDSFor;
