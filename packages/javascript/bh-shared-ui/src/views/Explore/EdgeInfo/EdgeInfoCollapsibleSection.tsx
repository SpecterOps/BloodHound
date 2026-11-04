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
import { faChevronDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Accordion, AccordionContent, AccordionHeader, AccordionItem, Typography } from 'doodle-ui';
import React, { PropsWithChildren } from 'react';
import { cn } from '../../../utils';
import { useCollapsibleSectionStyles } from '../InfoStyles';

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
            type='single'
            collapsible
            value={isExpanded ? 'section' : ''}
            onValueChange={(value) => onChange(value === 'section')}>
            <AccordionItem value='section' className='mb-0 bg-transparent'>
                <AccordionHeader
                    aria-label={label}
                    data-testid={`${label.toLocaleLowerCase()}-accordion`}
                    className='px-0 py-2 hover:no-underline'>
                    <FontAwesomeIcon icon={faChevronDown} aria-hidden className='mr-2 size-3 text-text-muted' />
                    <Typography component='span' variant='h6' className='text-sm'>
                        {label}
                    </Typography>
                </AccordionHeader>
                <AccordionContent
                    className={cn(
                        styles.edgeAccordionDetails,
                        '[&&]:px-0 [&&]:pb-3 [&&]:pt-0 [&_.typography-body2]:text-body-2'
                    )}>
                    {children}
                </AccordionContent>
            </AccordionItem>
        </Accordion>
    );
};

export default EdgeInfoCollapsibleSection;
