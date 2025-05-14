import { faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Accordion, AccordionDetails, AccordionSummary, Alert, AlertTitle } from '@mui/material';
import React, { PropsWithChildren } from 'react';
import { useCollapsibleSectionStyles } from '../InfoStyles';
import { SubHeader } from '../fragments';

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
        onChange?: (isOpen: boolean) => void;
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
                onChange(expanded);
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
