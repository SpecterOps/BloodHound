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
import React, { PropsWithChildren } from 'react';
import { useSelector } from 'react-redux';
import { EdgeInfoState, EdgeSections, edgeSectionToggle } from 'src/ducks/edgeinfo/edgeSlice';
import { AppState, useAppDispatch } from 'src/store';
import useCollapsibleSectionStyles from 'src/views/Explore/InfoStyles/CollapsibleSection';
import { SubHeader } from 'src/views/Explore/fragments';

export const EdgeInfoCollapsibleSection: React.FC<
    PropsWithChildren<{
        section: keyof typeof EdgeSections;
    }>
> = ({ children, section }) => {
    const styles = useCollapsibleSectionStyles();

    const dispatch = useAppDispatch();
    const edgeInfoState: EdgeInfoState = useSelector((state: AppState) => state.edgeinfo);

    const expanded = edgeInfoState.expandedSections[section];

    return (
        <Accordion
            expanded={expanded}
            onChange={() => dispatch(edgeSectionToggle({ section: section, expanded: !expanded }))}
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
