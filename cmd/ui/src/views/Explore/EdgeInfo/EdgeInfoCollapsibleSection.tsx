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
import { EdgeInfoState, EdgeSections, edgeSectionToggle, SubHeader, useCollapsibleSectionStyles } from 'bh-shared-ui';
import React, { PropsWithChildren } from 'react';
import { useAppDispatch, useAppSelector } from 'src/store';

export const EdgeInfoCollapsibleSection: React.FC<
    PropsWithChildren<{
        section: keyof typeof EdgeSections;
        onChange?: (label: string, isOpen: boolean) => void;
    }>
> = ({ children, section, onChange = () => {} }) => {
    const styles = useCollapsibleSectionStyles();

    const dispatch = useAppDispatch();
    const edgeInfoState: EdgeInfoState = useAppSelector((state) => state.edgeinfo);

    const expanded = edgeInfoState.expandedSections[section];

    return (
        <Accordion
            expanded={expanded}
            onChange={() => {
                dispatch(edgeSectionToggle({ section: section, expanded: !expanded }));
                onChange(section, !expanded);
            }}
            TransitionProps={{ unmountOnExit: true }}
            className={styles.accordionRoot}>
            <AccordionSummary
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
