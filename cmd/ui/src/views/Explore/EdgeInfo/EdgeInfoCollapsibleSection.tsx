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
import { SubHeader, useCollapsibleSectionStyles } from 'bh-shared-ui';
import React, { PropsWithChildren } from 'react';

export const EdgeInfoCollapsibleSection: React.FC<
    PropsWithChildren<{
        label: string;
        isExpanded: boolean;
        onChange?: (isOpen: boolean) => void;
    }>
> = ({ children, label = '', isExpanded, onChange = () => {} }) => {
    const styles = useCollapsibleSectionStyles();

    return (
        <Accordion
            expanded={isExpanded}
            onChange={(_e, expanded) => {
                onChange(expanded);
            }}
            TransitionProps={{ unmountOnExit: true }}
            className={styles.accordionRoot}>
            <AccordionSummary
                data-testid={`${label.toLocaleLowerCase()}-accordion`}
                expandIcon={<FontAwesomeIcon icon={isExpanded ? faMinus : faPlus} />}
                className={'accordion-summary'}
                classes={{
                    root: styles.accordionSummary,
                    expandIconWrapper: styles.expandIcon,
                }}>
                <SubHeader label={label} />
            </AccordionSummary>
            <AccordionDetails className={styles.edgeAccordionDetails}>{children}</AccordionDetails>
        </Accordion>
    );
};

export default EdgeInfoCollapsibleSection;
