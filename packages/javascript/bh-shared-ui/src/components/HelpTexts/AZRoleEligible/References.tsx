// Copyright 2025 Specter Ops, Inc.
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
                rel='noopener'
                href='https://learn.microsoft.com/en-us/rest/api/authorization/role-eligibility-schedule-instances/get?view=rest-authorization-2020-10-01&tabs=HTTP'>
                Role Eligibility Schedule Instances - Get - REST API (Azure Authorization)
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/graph/api/rbacapplication-list-roleeligibilityscheduleinstances?view=graph-rest-1.0&tabs=http'>
                List roleEligibilityScheduleInstances - Microsoft Graph v1.0
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/graph/api/policyroot-list-rolemanagementpolicies?view=graph-rest-1.0&tabs=http'>
                List roleManagementPolicies - Microsoft Graph v1.0
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/graph/api/policyroot-list-rolemanagementpolicyassignments?view=graph-rest-1.0&tabs=http'>
                List roleManagementPolicyAssignments - Microsoft Graph v1.0
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/graph/api/unifiedrolemanagementpolicyassignment-get?view=graph-rest-1.0&tabs=http'>
                Get unifiedRoleManagementPolicyAssignment - Microsoft Graph v1.0
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-apis'>
                API concepts in Privileged Identity management - Microsoft Entra ID Governance
            </Link>
        </Box>
    );
};

export default References;
