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

import { EntityInfoDataTableProps } from 'bh-shared-ui';
import { getRACFGroupCanSubmitAsQuery, RACF_GROUP_CAN_SUBMIT_AS_SECTION } from './groupMembers';
import { RACFRelationshipGroup, RACFRelationshipSection } from './RACFUserRelationships';

const outboundSections: RACFRelationshipSection[] = [
    {
        label: RACF_GROUP_CAN_SUBMIT_AS_SECTION,
        queryKey: 'racf-group-can-submit-as',
        getQuery: getRACFGroupCanSubmitAsQuery,
        fallbackKind: 'RACFUser',
    },
];

export const RACFGroupOutboundRelationships = (props: EntityInfoDataTableProps) => (
    <RACFRelationshipGroup {...props} relationshipSections={outboundSections} />
);
