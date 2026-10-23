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

import { Alert, Typography } from 'doodle-ui';
import React, { PropsWithChildren } from 'react';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';

const accordionCountClassName =
    'flex h-[1.6rem] min-w-12 items-center justify-center rounded-lg bg-neutral-5 px-2 text-[0.9rem] font-bold leading-[1.6em]';

export const Section: React.FC<PropsWithChildren<{ label?: string | null; className?: string }>> = ({
    label,
    className = '',
    children,
}) => {
    return (
        <div className={className}>
            {label && (
                <Typography variant='h6'>
                    <span
                        role='button'
                        aria-label={label}
                        tabIndex={0}
                        className={'link'}
                        onClick={(e) => {
                            e.preventDefault();
                        }}
                        onKeyDown={adaptClickHandlerToKeyDown((e) => e.preventDefault())}>
                        {label}
                    </span>
                </Typography>
            )}
            {children}
        </div>
    );
};

export const SubHeader: React.FC<{ label: string; count?: number; isLoading?: boolean; isError?: boolean }> = ({
    label,
    count,
    isLoading = false,
    isError = false,
}) => {
    return (
        <div className='flex w-full items-center justify-between'>
            <Typography variant='h6' className='ml-4 text-sm leading-[3em]'>
                {label}
            </Typography>
            {isLoading ? (
                <div className={accordionCountClassName}>
                    <span
                        aria-label={`Loading ${label}`}
                        className='size-5 animate-spin rounded-full border-2 border-neutral-4 border-t-primary'
                        role='progressbar'
                    />
                </div>
            ) : isError ? (
                <Alert
                    className='h-[1.6rem] w-12 items-center justify-center p-0 [&>div:first-child]:m-0 [&>div:nth-child(2)]:sr-only'
                    variant='error'>
                    Error loading {label}
                </Alert>
            ) : (
                count !== undefined && <span className={accordionCountClassName}>{count.toLocaleString()}</span>
            )}
        </div>
    );
};
