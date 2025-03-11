// Copyright 2023 Specter Ops, Inc.
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

import { faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Accordion, AccordionDetails, AccordionSummary } from '@mui/material';
import {
    EdgeInfoState,
    EdgeSections,
    SubHeader,
    edgeSectionToggle,
    useCollapsibleSectionStyles,
    useExploreParams,
    useFeatureFlag,
} from 'bh-shared-ui';
import React, { PropsWithChildren } from 'react';
import { useAppDispatch, useAppSelector } from 'src/store';

export const EdgeInfoCollapsibleSection: React.FC<
    PropsWithChildren<{
        section: keyof typeof EdgeSections;
        onChange?: (label: string, isOpen: boolean) => void;
    }>
> = ({ children, section, onChange = () => {} }) => {
    const styles = useCollapsibleSectionStyles();
    const { data: backButtonFlag } = useFeatureFlag('back_button_support');
    const { setExploreParams, expandedRelationships, selectedItem } = useExploreParams();

    const dispatch = useAppDispatch();
    const edgeInfoState: EdgeInfoState = useAppSelector((state) => state.edgeinfo);

    const setExpandedSection = () => {
        if (backButtonFlag?.enabled) {
            dispatch(
                edgeSectionToggle({
                    section: expandedRelationships?.at(0) as keyof typeof EdgeSections,
                    expanded: true,
                })
            );
        }
        return edgeInfoState.expandedSections[section];
    };

    const expanded = setExpandedSection();

    const setExpandedRelationshipsParam = () => {
        setExploreParams({
            expandedRelationships: [section],
            ...(section === 'composition' ? { searchType: 'composition' } : { searchType: null }),
            ...(section === 'composition' && { relationshipQueryItemId: selectedItem }),
        });
    };

    const handleOnChange = () => {
        if (backButtonFlag?.enabled) {
            dispatch(
                edgeSectionToggle({
                    section: expandedRelationships?.at(0) as keyof typeof EdgeSections,
                    expanded: false,
                })
            );
            setExpandedRelationshipsParam();
        } else {
            dispatch(edgeSectionToggle({ section: section, expanded: !expanded }));
        }
        onChange(section, !expanded);
    };

    return (
        <Accordion
            expanded={expanded}
            onChange={() => {
                handleOnChange();
            }}
            TransitionProps={{ unmountOnExit: true }}
            className={styles.accordionRoot}>
            <AccordionSummary
                data-testid={`${section.toLocaleLowerCase()}-accordion`}
                expandIcon={<FontAwesomeIcon icon={expanded ? faMinus : faPlus} />}
                className={'accordion-summary'}
                classes={{
                    root: styles.accordionSummary,
                    expandIconWrapper: styles.expandIcon,
                }}>
                <SubHeader label={EdgeSections[section]} />
            </AccordionSummary>
            <AccordionDetails className={styles.edgeAccordionDetails}>{children}</AccordionDetails>
        </Accordion>
    );
};

export default EdgeInfoCollapsibleSection;
