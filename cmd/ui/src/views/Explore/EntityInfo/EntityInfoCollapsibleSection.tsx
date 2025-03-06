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
import { Accordion, AccordionDetails, AccordionSummary, Alert, AlertTitle } from '@mui/material';
import { SubHeader, useCollapsibleSectionStyles } from 'bh-shared-ui';
import React, { PropsWithChildren } from 'react';

const EntityInfoCollapsibleSectionError: React.FC<{ error: any }> = ({ error }) => {
    //TODO: Once azure backend changes for counts param are in, utilize response error details
    let statusMessage = 'An unknown error occurred during the request.';
    switch (error?.response?.status) {
        case 500:
            statusMessage = 'The request could not be completed, possibly due to the volume of objects involved.';
            break;
        case 504:
            statusMessage = 'The results took too long to complete, possibly due to the volume of objects involved.';
            break;
        default:
            break;
    }
    return (
        <Alert severity='error' icon={false}>
            <AlertTitle sx={{ fontSize: '0.75rem' }}>{statusMessage}</AlertTitle>
        </Alert>
    );
};

export const EntityInfoCollapsibleSection: React.FC<
    PropsWithChildren<{
        label?: string;
        count?: number;
        onChange?: (label: string, isOpen: boolean) => void;
        isLoading?: boolean;
        isError?: boolean;
        error?: any;
        isExpanded: boolean;
    }>
> = ({
    children,
    label = '',
    count,
    onChange = () => {},
    isLoading = false,
    isError = false,
    error = null,
    isExpanded,
}) => {
    const styles = useCollapsibleSectionStyles();
    const disabled = isLoading || (count === 0 && !isError);

    return (
        <Accordion
            expanded={isExpanded}
            onChange={(_e, expanded) => {
                onChange(label, expanded);
            }}
            disabled={disabled}
            TransitionProps={{ unmountOnExit: true }}
            className={styles.accordionRoot}>
            <AccordionSummary
                expandIcon={<FontAwesomeIcon icon={isExpanded ? faMinus : faPlus} />}
                className={'accordion-summary'}
                classes={{
                    root: styles.accordionSummary,
                    expandIconWrapper: styles.expandIcon,
                }}>
                <SubHeader label={label} isLoading={isLoading} isError={isError} count={count} />
            </AccordionSummary>
            <AccordionDetails className={styles.accordionDetails}>
                {isError ? <EntityInfoCollapsibleSectionError error={error} /> : children}
            </AccordionDetails>
        </Accordion>
    );
};

export default EntityInfoCollapsibleSection;
