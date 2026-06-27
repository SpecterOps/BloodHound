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

import { EntityInfoCollapsibleSection, EntityInfoDataTableProps, useExploreParams } from 'bh-shared-ui';
import { useQueries } from 'react-query';
import {
    getRACFUserCanSubmitAsQuery,
    getRACFUserClassAuthoritiesQuery,
    getRACFUserSubmittedAsByQuery,
    RACF_USER_CAN_SUBMIT_AS_SECTION,
    RACF_USER_CLASS_AUTHORITIES_SECTION,
    RACF_USER_SUBMITTED_AS_BY_SECTION,
} from './groupMembers';
import { RACFRelatedNodes, RACFRelatedNodesProps } from './RACFGroupMembers';
import { fetchRACFRelatedNodes } from './relatedNodes';

export type RACFRelationshipSection = Pick<RACFRelatedNodesProps, 'label' | 'queryKey' | 'getQuery' | 'fallbackKind'>;

const outboundSections: RACFRelationshipSection[] = [
    {
        label: RACF_USER_CAN_SUBMIT_AS_SECTION,
        queryKey: 'racf-user-can-submit-as',
        getQuery: getRACFUserCanSubmitAsQuery,
        fallbackKind: 'RACFUser',
    },
    {
        label: RACF_USER_CLASS_AUTHORITIES_SECTION,
        queryKey: 'racf-user-class-authorities',
        getQuery: getRACFUserClassAuthoritiesQuery,
        fallbackKind: 'RACFClass',
    },
];

const inboundSections: RACFRelationshipSection[] = [
    {
        label: RACF_USER_SUBMITTED_AS_BY_SECTION,
        queryKey: 'racf-user-submitted-as-by',
        getQuery: getRACFUserSubmittedAsByQuery,
        fallbackKind: 'RACFUser',
    },
];

type RACFRelationshipGroupProps = EntityInfoDataTableProps & {
    relationshipSections: RACFRelationshipSection[];
};

export const RACFRelationshipGroup = ({
    id,
    label,
    parentLabels = [],
    relationshipSections,
}: RACFRelationshipGroupProps) => {
    const { expandedPanelSections, setExploreParams } = useExploreParams();
    const isExpanded = !!expandedPanelSections?.includes(label);
    const sectionQueries = useQueries(
        relationshipSections.map((section) => ({
            queryKey: [section.queryKey, id],
            queryFn: () => fetchRACFRelatedNodes(id, section.getQuery, section.fallbackKind),
            refetchOnWindowFocus: false,
            retry: false,
        }))
    );

    const count = sectionQueries.reduce((total, query) => total + (query.data?.length || 0), 0);
    const isLoading = sectionQueries.some((query) => query.isLoading);
    const failedQuery = sectionQueries.find((query) => query.isError);

    const handleOnChange = (isOpen: boolean) => {
        setExploreParams({
            expandedPanelSections: isOpen ? [...parentLabels, label] : parentLabels,
        });
    };

    return (
        <EntityInfoCollapsibleSection
            label={label}
            count={count}
            isExpanded={isExpanded}
            isLoading={isLoading}
            isError={!!failedQuery}
            error={failedQuery?.error}
            onChange={handleOnChange}>
            {relationshipSections.map((section) => (
                <RACFRelatedNodes
                    key={section.queryKey}
                    id={id}
                    label={section.label}
                    queryKey={section.queryKey}
                    getQuery={section.getQuery}
                    fallbackKind={section.fallbackKind}
                    parentLabels={[...parentLabels, label]}
                />
            ))}
        </EntityInfoCollapsibleSection>
    );
};

export const RACFUserOutboundRelationships = (props: EntityInfoDataTableProps) => (
    <RACFRelationshipGroup {...props} relationshipSections={outboundSections} />
);

export const RACFUserInboundRelationships = (props: EntityInfoDataTableProps) => (
    <RACFRelationshipGroup {...props} relationshipSections={inboundSections} />
);
